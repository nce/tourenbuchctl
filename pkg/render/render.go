package render

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type PageOpts struct {
	AbsoluteAssetDir string
	AbsoluteTextDir  string
	AbsoluteCwd      string
	RelativeCwd      string
	TmpDir           string
	ActivityName     string
	ActivityDate     string
	ActivityCategory string
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}

func (n *PageOpts) extractGpxData() error {
	fmt.Println(n.AbsoluteCwd)
	cmd := exec.Command("python3", n.AbsoluteAssetDir+"/meta/gpxplot.py", n.AbsoluteAssetDir+n.RelativeCwd+"/input.gpx")

	outfile, err := os.Create(n.TmpDir + "/gpxdata.txt")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to extract GPX data: %v", err)
	}
	return nil
}

func (n *PageOpts) generatElevationProfile() error {
	fmt.Println(n.TmpDir)
	cmd := exec.Command("gnuplot", "elevation.plt")
	cmd.Dir = n.AbsoluteTextDir + n.RelativeCwd + "/" + n.TmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to generate elevation profile: %v", err)
	}
	return nil
}

func (n *PageOpts) generateLatexDescription() error {
	cmd := exec.Command("pandoc", "--from", "markdown+tex_math_dollars", "--variable=assetdir:"+n.AbsoluteAssetDir+"/"+n.RelativeCwd, "--variable=path:.", "--variable=projectroot:"+n.AbsoluteTextDir, "--template", n.AbsoluteTextDir+"meta/tourenbuch.template", "--metadata-file", n.AbsoluteCwd+"/header.yaml", n.AbsoluteCwd+"/description.md", "--output", n.TmpDir+"/description.tex")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to generate elevation profile: %v", err)
	}
	return nil

}

type category struct {
	category string `yaml:"type"`
}

type activity struct {
	activity category `yaml:"activity"`
}

func getActivityTypeFromHeader() (string, error) {
	headerFile := "header.yaml"

	yamlfile, err := os.ReadFile(headerFile)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	var act activity

	err = yaml.Unmarshal(yamlfile, &act)
	if nil != err {
		return "", fmt.Errorf("error unmarshalling yaml: %w", err)
	}

	return act.activity.category, nil
}

func NewPage(cwd string) (*PageOpts, error) {

	return &PageOpts{
		AbsoluteAssetDir: "/Users/nce/Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/",
		AbsoluteTextDir:  "/Users/nce/vcs/github/nce/tourenbuch/",
		AbsoluteCwd:      cwd,
		RelativeCwd:      strings.TrimPrefix(cwd, "/Users/nce/vcs/github/nce/tourenbuch/"),
	}, nil
}

func (n *PageOpts) GenerateSinglePageActivity() error {
	// Step 1: Create a temporary directory
	tempDir, err := os.MkdirTemp(".", "tmp")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	//defer os.RemoveAll(tempDir)

	n.TmpDir = tempDir

	// Step 2: Store the LaTeX file in the temporary directory
	latexFilePath := filepath.Join(tempDir, "document.tex")

	err = os.WriteFile(latexFilePath, []byte(latexContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write LaTeX file: %v", err)
	}

	err = copyFile("elevation.plt", filepath.Join(tempDir, "elevation.plt"))
	if err != nil {
		log.Fatalf("Failed to copy LaTeX file: %v", err)
	}

	n.extractGpxData()
	n.generateLatexDescription()
	n.generatElevationProfile()

	// Step 3: Compile the LaTeX file with pdflatex
	cmd := exec.Command("pdflatex", "-shell-escape", "-output-directory", tempDir, latexFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to compile LaTeX file: %v", err)
	}

	// Step 4: Open the generated PDF with a PDF viewer
	pdfFilePath := filepath.Join(tempDir, "document.pdf")
	viewerCmd := exec.Command("open", pdfFilePath) // Use "open" on macOS or "start" on Windows
	viewerCmd.Stdout = os.Stdout
	viewerCmd.Stderr = os.Stderr
	err = viewerCmd.Run()
	if err != nil {
		log.Fatalf("Failed to open PDF file: %v", err)
	}

	fmt.Println("PDF generated and opened successfully.")

	return nil
}
