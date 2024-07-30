package render

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/nce/tourenbuchctl/pkg/utils"
)

type PageOpts struct {
	AbsoluteAssetDir string
	AbsoluteCwd      string
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
	cmd := exec.Command("python3", n.AbsoluteAssetDir, n.AbsoluteAssetDir+n.ActivityCategory+"/"+n.ActivityName+"-"+n.ActivityDate+"/input.gpx")

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

func generatElevationProfile(tmpDir string) error {
	cmd := exec.Command("gnuplot", "elevation.plt")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to generate elevation profile: %v", err)
	}
	return nil
}

func generateLatexDescription() error {

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

	dir, date, err := utils.SplitActivityDirectoryName(cwd)
	if err != nil {
		return nil, fmt.Errorf("error splitting activity directory name: %w", err)
	}

	category, err := getActivityTypeFromHeader()
	if err != nil {
		return nil, fmt.Errorf("error getting activity type from header: %w", err)
	}

	return &PageOpts{
		AbsoluteAssetDir: "/Users/nce/Library/Mobile Documents/com~apple~CloudDocs/privat/sport/Tourenbuch/",
		AbsoluteCwd:      cwd,
		ActivityName:     dir,
		ActivityDate:     date,
		ActivityCategory: category,
	}, nil
}

func (n *PageOpts) GenerateSinglePageActivity() error {
	// Step 1: Create a temporary directory
	tempDir, err := os.MkdirTemp(".", "tmp")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	n.TmpDir = tempDir

	// Step 2: Store the LaTeX file in the temporary directory
	latexFilePath := filepath.Join(tempDir, "document.tex")

	err = os.WriteFile(latexFilePath, []byte(latexContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write LaTeX file: %v", err)
	}

	err = copyFile("description.tex", filepath.Join(tempDir, "description.tex"))
	if err != nil {
		log.Fatalf("Failed to copy LaTeX file: %v", err)
	}

	n.extractGpxData()
	n.generateLatexDescription()

	// Step 3: Compile the LaTeX file with pdflatex
	cmd := exec.Command("pdflatex", "-output-directory", tempDir, latexFilePath)
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
