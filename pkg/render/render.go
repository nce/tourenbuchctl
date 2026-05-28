package render

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"github.com/nce/tourenbuchctl/pkg/maprender"
	"github.com/nce/tourenbuchctl/pkg/pdfexport"
	"github.com/nce/tourenbuchctl/pkg/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type PageOpts struct {
	AbsoluteAssetDir     string
	AbsoluteTextDir      string
	AbsoluteCwd          string
	RelativeCwd          string
	TmpDir               string
	ActivityName         string
	ActivityDate         string
	ActivityType         string
	ElevationProfileType string
	ExportToDisk         bool
	ExportToS3           bool
	MaxElevation         string
	ActivityTitle        string
	Compression          bool
}

const (
	relativeTextLibraryPath  = "vcs/github/nce/tourenbuch/"
	relativeAssetLibraryPath = "Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/"
)

const latexContent = `
\input{/Users/nce/` + relativeTextLibraryPath + `/meta/header}
\begin{document}
\include{description.tex}
\end{document}
`

//go:embed templates/*
var content embed.FS

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer input.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, input)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	err = out.Close()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	return nil
}

func NewPage(cwd string, exportToDisk bool, exportToS3 bool, compression bool) (*PageOpts, error) {
	name, date, err := utils.SplitActivityDirectoryName(filepath.Base(cwd))
	if err != nil {
		return nil, fmt.Errorf("failed to split activity directory name: %w", err)
	}

	activity, err := activity.GetFromHeader[string](cwd,
		"Activity.Type",
		"Layout.ElevationProfileType",
		"Activity.MaxElevation",
		"Activity.Title")
	if err != nil {
		return nil, fmt.Errorf("failed to read values from header.yaml: %w", err)
	}

	return &PageOpts{
		AbsoluteAssetDir:     "/Users/nce/Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/",
		AbsoluteTextDir:      "/Users/nce/vcs/github/nce/tourenbuch/",
		AbsoluteCwd:          cwd,
		RelativeCwd:          strings.TrimPrefix(cwd, "/Users/nce/vcs/github/nce/tourenbuch/"),
		ExportToDisk:         exportToDisk,
		ExportToS3:           exportToS3,
		ActivityName:         name,
		ActivityDate:         date,
		ActivityType:         activity["Activity.Type"],
		ElevationProfileType: activity["Layout.ElevationProfileType"],
		MaxElevation:         activity["Activity.MaxElevation"],
		ActivityTitle:        activity["Activity.Title"],
		Compression:          compression,
	}, nil
}

func (n *PageOpts) GenerateSinglePageActivity(preventCleanup bool) error {
	tempDir, err := os.MkdirTemp(".", "tmp")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create temp directory")
	}

	defer func() {
		if !preventCleanup {
			time.Sleep(1 * time.Second) // Sleep for 1 second before removing the directory
			os.RemoveAll(tempDir)
		} else {
			log.Info().Msgf("Asset rendering folder not removed (%s)", tempDir)
		}
	}()

	n.TmpDir = tempDir

	latexFilePath := filepath.Join(tempDir, "document.tex")

	//nolint: gosec
	err = os.WriteFile(latexFilePath, []byte(latexContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write LaTeX file: %w", err)
	}

	if err = n.extractGpxData(); err != nil {
		return fmt.Errorf("failed to extract gpx data: %w", err)
	}

	if err = n.generateMap(); err != nil {
		return fmt.Errorf("failed to generate map: %w", err)
	}

	if err = n.generateLatexDescription(); err != nil {
		return fmt.Errorf("failed to generate latex description: %w", err)
	}

	if err = n.generatElevationProfile(); err != nil {
		return fmt.Errorf("failed to generate elevation profile: %w", err)
	}

	cmd := exec.CommandContext(
		context.Background(),
		"pdflatex",
		"-shell-escape",
		"-output-directory", tempDir, latexFilePath,
	)

	var stdout bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w\n%s", err, stdout.String())
	}

	if n.Compression {
		//nolint: gosec
		cmd := exec.CommandContext(
			context.Background(),
			"gs",
			"-sDEVICE=pdfwrite",
			"-dCompatibilityLevel=1.4",
			"-dPDFSETTINGS=/ebook",
			"-dNOPAUSE -dQUIET -dBATCH",
			fmt.Sprintf("-sOutputFile=%s/compressed.pdf", tempDir),
			tempDir+"/document.pdf",
		)

		var stdout bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to compress PDF: %w\n%s", err, stdout.String())
		}

		err = os.Rename(tempDir+"/compressed.pdf", tempDir+"/document.pdf")
		if err != nil {
			return fmt.Errorf("failed to rename compressed PDF: %w", err)
		}
	}

	if n.ExportToDisk {
		local := pdfexport.LocalExport{
			DestDirectory: n.AbsoluteAssetDir + n.RelativeCwd,
			DestFilename:  n.ActivityDate + "-" + n.ActivityDate + ".pdf",
		}

		if err := local.Save(tempDir + "/document.pdf"); err != nil {
			return fmt.Errorf("failed to export to Diskpath: %s; %w", local.DestDirectory, err)
		}

		log.Info().Str("export", local.DestDirectory).Msg("PDF saved to disk")
	}

	if n.ExportToS3 {
		//nolint: varnamelen
		s3 := pdfexport.S3Export{
			BucketName: "tourenbuch",
			ObjectName: n.ActivityType + "/" + n.ActivityName + "-" + n.ActivityDate + ".pdf",
		}

		if err := s3.Save(tempDir + "/document.pdf"); err != nil {
			return fmt.Errorf("failed to export to s3: %s; %w", s3.ObjectName, err)
		}

		log.Info().
			Str("bucket", s3.BucketName).
			Str("object", s3.ObjectName).
			Msg("PDF saved to s3")
	}

	pdfFilePath := filepath.Join(tempDir, "document.pdf")
	viewerCmd := exec.CommandContext(
		context.Background(),
		"open",
		pdfFilePath,
	)

	viewerCmd.Stdout = os.Stdout
	viewerCmd.Stderr = os.Stderr

	err = viewerCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}

	return nil
}

func (n *PageOpts) extractGpxData() error {
	//nolint: gosec
	cmd := exec.CommandContext(
		context.Background(),
		"python3",
		n.AbsoluteAssetDir+"/meta/gpxplot.py",
		n.AbsoluteAssetDir+n.RelativeCwd+"/input.gpx",
	)

	outfile, err := os.Create(n.TmpDir + "/gpxdata.txt")
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to extract gpx data: %w", err)
	}

	return nil
}

func (n *PageOpts) generateMap() error {
	apiKey := viper.GetString("THUNDERFOREST_API_KEY")
	if apiKey == "" {
		log.Info().Msg("THUNDERFOREST_API_KEY not configured; skipping map generation")

		return nil
	}

	inputPath := filepath.Join(n.AbsoluteAssetDir, n.RelativeCwd, "input.gpx")
	outputPath := filepath.Join(n.AbsoluteAssetDir, n.RelativeCwd, "map.png")

	if _, err := os.Stat(outputPath); err == nil {
		log.Info().Str("mapFile", outputPath).Msg("Map already exists; skipping map generation")

		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking map file: %w", err)
	}

	err := maprender.GenerateForActivity(context.Background(), inputPath, outputPath, apiKey, n.ActivityType)
	if err != nil {
		if errors.Is(err, maprender.ErrMissingAPIKey) {
			log.Info().Msg("THUNDERFOREST_API_KEY not configured; skipping map generation")

			return nil
		}

		return fmt.Errorf("generating thunderforest map: %w", err)
	}

	log.Info().Str("mapFile", outputPath).Msg("Generated Thunderforest map")

	return nil
}

func (n *PageOpts) prepareElevationProfile() error {
	absoluteTempDir := n.AbsoluteTextDir + n.RelativeCwd + "/" + n.TmpDir + "/"

	err := copyFile("elevation.plt", filepath.Join(n.TmpDir, "elevation.plt"))
	if err != nil {
		return fmt.Errorf("failed to copy Elevation label file: %w", err)
	}

	settingsLocation := "templates/elevationprofile/activity-settings.plt"

	tmpl, err := template.ParseFS(content, settingsLocation)
	if err != nil {
		return fmt.Errorf("parsing template file: %s; error: %w", settingsLocation, err)
	}

	file, err := os.Create(absoluteTempDir + "activity-settings.plt")
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, nil)
	if err != nil {
		return fmt.Errorf("writing template to file: %w", err)
	}

	elevationProfile := fmt.Sprintf("templates/elevationprofile/%s.plt", n.ElevationProfileType)

	tmpl, err = template.ParseFS(content, elevationProfile)
	if err != nil {
		return fmt.Errorf("parsing template file: %s; error: %w", elevationProfile, err)
	}

	file, err = os.Create(absoluteTempDir + "master.plt")
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer file.Close()

	data := struct {
		Activity string
	}{
		Activity: n.ActivityType,
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("writing template to file: %w", err)
	}

	return nil
}

func (n *PageOpts) generatElevationProfile() error {
	if err := n.prepareElevationProfile(); err != nil {
		return fmt.Errorf("failed to prepare elevation profile: %w", err)
	}

	//nolint: gosec
	cmd := exec.CommandContext(
		context.Background(),
		"gnuplot",
		"-c",
		"master.plt",
		n.ActivityTitle,
		n.MaxElevation,
	)

	cmd.Dir = n.AbsoluteTextDir + n.RelativeCwd + "/" + n.TmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate elevation profile: %w", err)
	}

	return nil
}

func (n *PageOpts) generateLatexDescription() error {
	//nolint: gosec
	cmd := exec.CommandContext(
		context.Background(),
		"pandoc",
		"--from", "markdown+tex_math_dollars",
		"--variable=assetdir:"+n.AbsoluteAssetDir+"/"+n.RelativeCwd,
		"--variable=path:.",
		"--variable=projectroot:"+n.AbsoluteTextDir,
		"--variable=omitPageNumber:true",
		"--template", n.AbsoluteTextDir+"meta/tourenbuch.template",
		"--metadata-file", n.AbsoluteCwd+"/header.yaml",
		n.AbsoluteCwd+"/description.md",
		"--output", n.TmpDir+"/description.tex",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate latex description: %w", err)
	}

	return nil
}
