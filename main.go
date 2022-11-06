package main

import (
	"embed"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed .eslintrc.cjs
var eslintrc string

//go:embed .prettierrc.cjs
var prettierrc string

//go:embed tailwind.config.cjs
var tailwindConfig string

//go:embed src
var src embed.FS

type configuration struct {
	projectName     string
	destinationPath string
}

type application struct {
	cfg configuration
}

func copyDir(root string, destination string) {
	println(root)
	currentDir, err := src.ReadDir(root)

	if err != nil {
		panic(err)
	}

	for _, d := range currentDir {
		if d.IsDir() {
			nextDestination := filepath.Join(destination, d.Name())
			_ = os.MkdirAll(nextDestination, 0755)
			copyDir(filepath.Join(root, d.Name()), nextDestination)
			continue
		}

		sourceFilePath := filepath.Join(root, d.Name())

		f, fErr := src.ReadFile(sourceFilePath)
		if fErr != nil {
			log.Fatal(fErr)
		}

		destinationFilePath := filepath.Join(destination, d.Name())

		writeFileErr := os.WriteFile(destinationFilePath, f, 0755)

		if writeFileErr != nil {
			log.Fatal(writeFileErr)
		}

		fmt.Printf("Copied %s to %s\n", sourceFilePath, destinationFilePath)
	}
}

func main() {
	var cfg configuration
	flag.StringVar(&cfg.projectName, "project", "", "project name")
	flag.Parse()

	if cfg.projectName == "" {
		flag.PrintDefaults()
		return
	}

	workingDir, wdErr := os.Getwd()
	if wdErr != nil {
		panic(wdErr)
	}

	destinationPath := filepath.Join(workingDir, cfg.projectName)

	if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
		cfg.destinationPath = destinationPath
		app := application{cfg}
		app.runVite()
		app.chDir()
		app.installDependencies()
		app.installESLint()
		app.installTailwind()
		app.src()
	} else {
		panic("Project directory already exists")
	}
}

func (app *application) runVite() {
	// TODO make this work with different templates
	cmd := exec.Command("npm", "create", "vite@latest", app.cfg.projectName, "--", "--template", "react-ts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()

	if cmdErr != nil {
		log.Fatal(cmdErr)
	}
}

func (app *application) chDir() {
	chDirErr := os.Chdir(app.cfg.destinationPath)
	if chDirErr != nil {
		log.Fatal(chDirErr)
	}
}

func (app *application) installDependencies() {
	dependencies := []string{
		"react-router-dom",
		"react-use",
		"zustand",
	}

	app.runNPMInstall(false, dependencies...)
}

func (app *application) installESLint() {
	devDeps := []string{
		"eslint",
		"eslint-config-prettier",
		"eslint-plugin-react",
		"prettier",
		"prettier-plugin-tailwindcss",
		"@typescript-eslint/eslint-plugin",
		"@typescript-eslint/parser",
		"@trivago/prettier-plugin-sort-imports",
		"@types/node",
	}

	app.runNPMInstall(true, devDeps...)

	eslintRC, eslintErr := os.Create(".eslintrc.cjs")
	if eslintErr != nil {
		log.Fatal(eslintErr)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(eslintRC)

	_, copyErr := io.Copy(eslintRC, strings.NewReader(eslintrc))

	if copyErr != nil {
		log.Fatal(copyErr)
	}

	prettierRC, prettierErr := os.Create(".prettierrc.cjs")

	if prettierErr != nil {
		log.Fatal(prettierErr)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(prettierRC)

	_, copyErr = io.Copy(prettierRC, strings.NewReader(prettierrc))

	if copyErr != nil {
		log.Fatal(copyErr)
	}
}

func (app *application) src() {
	srcPath := filepath.Join(app.cfg.destinationPath, "src")
	removeAllErr := os.RemoveAll(srcPath)

	if removeAllErr != nil {
		log.Fatal(removeAllErr)
	}

	mkdirErr := os.MkdirAll(srcPath, 0755)

	if mkdirErr != nil {
		log.Fatal(mkdirErr)
	}

	copyDir("src", "src")

}

func (app *application) installTailwind() {
	devDeps := []string{
		"tailwindcss",
		"postcss",
		"autoprefixer",
		"@tailwindcss/typography",
		"@tailwindcss/forms",
		"@tailwindcss/line-clamp",
		"@tailwindcss/aspect-ratio",
		"daisyui",
	}

	app.runNPMInstall(true, devDeps...)

	cmd := exec.Command("npx", "tailwindcss", "init", "-p")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		log.Fatal(cmdErr)
	}

}

func (app *application) runNPMInstall(isDev bool, libraries ...string) {
	var args []string

	if isDev {
		args = append([]string{"install", "-D"}, libraries...)
	} else {
		args = append([]string{"install"}, libraries...)
	}

	cmd := exec.Command("npm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()
	if cmdErr != nil {
		log.Fatal(cmdErr)
	}

	tf, tfErr := os.Create("tailwind.config.cjs")
	if tfErr != nil {
		log.Fatal(tfErr)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(tf)

	_, copyErr := io.Copy(tf, strings.NewReader(tailwindConfig))

	if copyErr != nil {
		log.Fatal(copyErr)
	}
}
