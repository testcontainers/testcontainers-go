package core

import (
	"bufio"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	IndexDockerIO   = "https://index.docker.io/v1/"
	maxURLRuneCount = 2083
	minURLRuneCount = 3
	URLSchema       = `((ftp|tcp|udp|wss?|https?):\/\/)`
	URLUsername     = `(\S+(:\S*)?@)`
	URLIP           = `([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))`
	IP              = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	URLSubdomain    = `((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))`
	URLPath         = `((\/|\?|#)[^\s]*)`
	URLPort         = `(:(\d{1,5}))`
	URL             = `^` + URLSchema + `?` + URLUsername + `?` + `((` + URLIP + `|(\[` + IP + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + URLSubdomain + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + URLPort + `?` + URLPath + `?$`
)

var rxURL = regexp.MustCompile(URL)

// ExtractImagesFromDockerfile extracts images from the Dockerfile sourced from dockerfile.
func ExtractImagesFromDockerfile(dockerfile string, buildArgs map[string]*string) ([]string, error) {
	file, err := os.Open(dockerfile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ExtractImagesFromReader(file, buildArgs)
}

// ExtractImagesFromReader extracts images from the Dockerfile sourced from r.
func ExtractImagesFromReader(r io.Reader, buildArgs map[string]*string) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	images := make([]string, 0, len(lines))

	// Track global ARGs declared before the first FROM
	globalArgs := make(map[string]string)
	for k, v := range buildArgs {
		if v != nil {
			globalArgs[k] = *v
		}
	}

	inGlobalScope := true

	// extract images from dockerfile
	for _, line := range lines {
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)

		if strings.HasPrefix(upper, "ARG ") || upper == "ARG" {
			if inGlobalScope {
				argRest := strings.TrimSpace(line[3:])
				name, defaultValue, ok := strings.Cut(argRest, "=")
				if ok {
					name = strings.TrimSpace(name)
					defaultValue = strings.Trim(strings.TrimSpace(defaultValue), "\"'")
					for k, v := range globalArgs {
						defaultValue = strings.ReplaceAll(defaultValue, "${"+k+"}", v)
						defaultValue = strings.ReplaceAll(defaultValue, "$"+k, v)
					}
					if override, exists := buildArgs[name]; !exists || override == nil {
						globalArgs[name] = defaultValue
					}
				}
			}
			continue
		}

		if !strings.HasPrefix(upper, "FROM") {
			continue
		}

		inGlobalScope = false

		// remove FROM
		line = strings.TrimSpace(line[4:])
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		// Skip optional FROM flags, e.g. --platform=linux/amd64
		imageIdx := 0
		for imageIdx < len(parts) && strings.HasPrefix(parts[imageIdx], "--") {
			imageIdx++
		}
		if imageIdx >= len(parts) {
			continue
		}

		imageName := parts[imageIdx]

		// interpolate global build args and defaults
		for k, v := range globalArgs {
			imageName = strings.ReplaceAll(imageName, "${"+k+"}", v)
			imageName = strings.ReplaceAll(imageName, "$"+k, v)
		}
		images = append(images, imageName)
	}

	return images, nil
}

// ExtractRegistry extracts the registry from the image name, using a regular expression to extract the registry from the image name.
// regular expression to extract the registry from the image name
// the regular expression is based on the grammar defined in
// - image:tag
// - image
// - repository/image:tag
// - repository/image
// - registry/image:tag
// - registry/image
// - registry/repository/image:tag
// - registry/repository/image
// - registry:port/repository/image:tag
// - registry:port/repository/image
// - registry:port/image:tag
// - registry:port/image
// Once extracted the registry, it is validated to check if it is a valid URL or an IP address.
func ExtractRegistry(image string, fallback string) string {
	exp := regexp.MustCompile(`^(?:(?P<registry>(https?://)?[^/]+)(?::(?P<port>\d+))?/)?(?:(?P<repository>[^/]+)/)?(?P<image>[^:]+)(?::(?P<tag>.+))?$`).FindStringSubmatch(image)
	if len(exp) == 0 {
		return ""
	}

	registry := exp[1]

	// docker.io is an implicit reference, return fallback for normalization
	if strings.EqualFold(registry, "docker.io") {
		return fallback
	}

	// registry.hub.docker.com is an explicit registry reference, preserve it
	if strings.EqualFold(registry, "registry.hub.docker.com") {
		return "registry.hub.docker.com"
	}

	if IsURL(registry) {
		return registry
	}

	return fallback
}

// IsURL checks if the string is a URL.
// Extracted from https://github.com/asaskevich/govalidator/blob/f21760c49a8d/validator.go#L104
func IsURL(str string) bool {
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false
	}
	strTemp := str
	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used so it does not impact rxURL.MatchString
		strTemp = "http://" + str
	}
	u, err := url.Parse(strTemp)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}
	return rxURL.MatchString(str)
}
