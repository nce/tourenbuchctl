package render

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"github.com/nce/tourenbuchctl/pkg/utils"
	"github.com/rs/zerolog/log"
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
	SaveToDisk           bool
	MaxElevation         string
	ActivityTitle        string
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

func (n *PageOpts) extractGpxData() error {
	//nolint: gosec
	cmd := exec.Command(
		"python3",
		n.AbsoluteAssetDir+"/meta/gpxplot.py",
		n.AbsoluteAssetDir+n.RelativeCwd+"/input.gpx")

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
	cmd := exec.Command(
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
	cmd := exec.Command(
		"pandoc",
		"--from", "markdown+tex_math_dollars",
		"--variable=assetdir:"+n.AbsoluteAssetDir+"/"+n.RelativeCwd,
		"--variable=path:.",
		"--variable=projectroot:"+n.AbsoluteTextDir,
		"--variable=omitPageNumber:true",
		"--template", n.AbsoluteTextDir+"meta/tourenbuch.template",
		"--metadata-file", n.AbsoluteCwd+"/header.yaml",
		n.AbsoluteCwd+"/description.md",
		"--output", n.TmpDir+"/description.tex")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate latex description: %w", err)
	}

	return nil
}

func NewPage(cwd string, saveToDisk bool) (*PageOpts, error) {
	name, date, err := utils.SplitActivityDirectoryName(filepath.Base(cwd))
	if err != nil {
		return nil, fmt.Errorf("failed to split activity directory name: %w", err)
	}

	activityType, err := activity.GetFromHeader[string](cwd, "Activity.Type")
	if err != nil {
		return nil, fmt.Errorf("failed to read activity type: %w", err)
	}

	elevationProfileType, err := activity.GetFromHeader[string](cwd, "Layout.ElevationProfileType")
	if err != nil {
		return nil, fmt.Errorf("failed to read elevationProfileType: %w", err)
	}

	maxElevation, err := activity.GetFromHeader[string](cwd, "Activity.MaxElevation")
	if err != nil {
		return nil, fmt.Errorf("failed to read activity type: %w", err)
	}

	activityTitle, err := activity.GetFromHeader[string](cwd, "Activity.Title")
	if err != nil {
		return nil, fmt.Errorf("failed to read activity type: %w", err)
	}

	return &PageOpts{
		AbsoluteAssetDir:     "/Users/nce/Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/",
		AbsoluteTextDir:      "/Users/nce/vcs/github/nce/tourenbuch/",
		AbsoluteCwd:          cwd,
		RelativeCwd:          strings.TrimPrefix(cwd, "/Users/nce/vcs/github/nce/tourenbuch/"),
		SaveToDisk:           saveToDisk,
		ActivityName:         name,
		ActivityDate:         date,
		ActivityType:         activityType,
		ElevationProfileType: elevationProfileType,
		MaxElevation:         maxElevation,
		ActivityTitle:        activityTitle,
	}, nil
}

func (n *PageOpts) GenerateSinglePageActivity() error {
	tempDir, err := os.MkdirTemp(".", "tmp")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create temp directory")
	}

	defer func() {
		time.Sleep(1 * time.Second) // Sleep for 1 second before removing the directory
		os.RemoveAll(tempDir)
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

	if err = n.generateLatexDescription(); err != nil {
		return fmt.Errorf("failed to generate latex description: %w", err)
	}

	if err = n.generatElevationProfile(); err != nil {
		return fmt.Errorf("failed to generate elevation profile: %w", err)
	}

	cmd := exec.Command(
		"pdflatex",
		"-shell-escape",
		"-output-directory", tempDir, latexFilePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	if n.SaveToDisk {
		storagePath := n.AbsoluteAssetDir + n.RelativeCwd + "/" +
			n.ActivityName + "-" + n.ActivityDate + ".pdf"

		err = copyFile(tempDir+"/document.pdf", storagePath)
		if err != nil {
			return fmt.Errorf("failed to save PDF to disk: %w", err)
		}

		log.Info().Str("storage", storagePath).Msg("PDF saved to disk")
	}

	pdfFilePath := filepath.Join(tempDir, "document.pdf")
	viewerCmd := exec.Command("open", pdfFilePath)

	viewerCmd.Stdout = os.Stdout
	viewerCmd.Stderr = os.Stderr

	err = viewerCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}

	return nil
}
