package config

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/Cycloctane/dnspass/internal/records"
)

func ParseFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	bufio.NewScanner(f)
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		split := strings.SplitN(line, "/", 2)
		if len(split) != 2 {
			continue
		}
		log.Printf("Adding record: %s -> %s\n", split[0], split[1])
		records.Add(split[0], split[1])
	}
	return nil
}
