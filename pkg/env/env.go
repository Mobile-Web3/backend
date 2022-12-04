package env

import (
	"bufio"
	"os"
	"strings"
)

func Parse() error {
	file, err := os.OpenFile(".env", os.O_RDONLY, 0777)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return err
		}

		line := scanner.Text()
		values := strings.Split(line, "=")
		if err = os.Setenv(values[0], values[1]); err != nil {
			return err
		}
	}

	return nil
}
