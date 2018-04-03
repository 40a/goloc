package goloc

import (
	"os"
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"path"
)

func WriteLocalizations(
	platform Platform,
	dir ResDir,
	localizations Localizations,
	defLocLang string,
	defLocPath string,
) error {
	// Make sure the the resources dir exists
	file, err := os.Open(dir)
	if err != nil {
		return err
	}

	writers := map[Lang]*bufio.Writer{}

	// For each localization: create a writer if needed and write each localized string
	for key, keyLoc := range localizations {
		for lang, value := range keyLoc {
			if writer, ok := writers[lang]; !ok { // Create a new writer and write a header to it if needed
				var resDir string
				var fileName string

				// Handle default language
				if len(defLocLang) > 0 && lang == defLocLang && len(defLocPath) > 0 {
					resDir = path.Dir(defLocPath)
					fileName = path.Base(defLocPath)
				} else {
					resDir = platform.LocalizationDirPath(lang, dir)
					fileName = platform.LocalizationFileName(lang)
				}

				// Create all intermediate directories
				err := os.MkdirAll(resDir, os.ModePerm)
				if err != nil {
					return err
				}

				// Create actual localization file
				file, err = os.Create(filepath.Join(resDir, fileName))
				// noinspection GoDeferInLoop
				defer file.Close()
				if err != nil {
					return err
				}

				// Open a new writer for the localization file
				writer := bufio.NewWriter(file)
				if writer == nil {
					return errors.New(fmt.Sprintf(`can't create a bufio.Writer for %v`, file))
				}

				writers[lang] = writer

				// Write a header
				_, err = writer.WriteString(platform.Header(lang))
				if err != nil {
					return errors.New(fmt.Sprintf(`can't write header to %v, reason: %v`, file, err))
				}
			} else { // Use an existing writer to write another localized string
				localizedString := platform.Localization(lang, key, value)
				if len(localizedString) < 1 {
					return errors.New(fmt.Sprintf(`can't write a new line to %v, reason: %v`, file, err))
				}

				_, err := writer.WriteString(localizedString)
				if err != nil {
					return errors.New(fmt.Sprintf(`can't write a new line to %v, reason: %v`, file, err))
				}
			}
		}
	}

	// For each writer: write a footer and flush everything
	for lang, writer := range writers {
		_, err := writer.WriteString(platform.Footer(lang))
		if err != nil {
			return errors.New(fmt.Sprintf(`can't write footer to %v, reason: %v`, file, err))
		}

		err = writer.Flush()
		if err != nil {
			return errors.New(fmt.Sprintf(`can't write to %v, reason: %v`, file, err))
		}
	}

	return nil
}