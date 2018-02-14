package helpers

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/xtgo/set"
)

const (
	GomeetDefaultPrefixes = "svc-,gomeet-svc-"
)

type Empty struct{}

type PkgNfo struct {
	goPkg             string
	name              string
	path              string
	shortName         string
	prefix            string
	defaultPrefixes   string
	projectGroupGoPkg string
	projectGroupName  string
}

type LogType int

const (
	LogError     LogType = iota - 1 // -1
	LogDangerous                    // 0
	LogSkipping                     // 1
	LogReplacing                    // 2
	LogCreating                     // 3
	LogInfo                         // 4
)

func Log(t LogType, msg string) {
	var p, head string
	switch t {
	case LogError:
		p, head = "%s     - %s\n", color.RedString("[Error]")
	case LogDangerous:
		p, head = "%s - %s\n", color.RedString("[Dangerous]")
	case LogSkipping:
		p, head = "%s  - %s\n", color.YellowString("[Skipping]")
	case LogReplacing:
		p, head = "%s - %s\n", color.YellowString("[Replacing]")
	case LogCreating:
		p, head = "%s  - %s\n", color.GreenString("[Creating]")
	case LogInfo:
		p, head = "%s      - %s\n", color.CyanString("[Info]")
	default:
		p, head = "%s    - %s\n", "[Unknow]"
	}

	log.Printf(p, head, msg)
}

func NewPkgNfo(goPkg, defaultPrefixes string) (*PkgNfo, error) {
	pNfo := &PkgNfo{}
	if err := pNfo.setGoPkg(goPkg); err != nil {
		return nil, err
	}
	if err := pNfo.SetDefaultPrefixes(defaultPrefixes); err != nil {
		return nil, err
	}

	return pNfo, nil
}

func (pNfo *PkgNfo) setGoPkg(goPkg string) (err error) {
	if pNfo.path, err = Path(goPkg); err != nil {
		return err
	}

	pNfo.goPkg = goPkg
	pNfo.name = strings.ToLower(LastFromSplit(pNfo.GoPkg(), "/"))
	if err = pNfo.SetDefaultPrefixes(pNfo.DefaultPrefixes()); err != nil {
		return err
	}

	pNfo.projectGroupGoPkg = filepath.Dir(pNfo.GoPkg())
	splitProjectGoPkg := strings.Split(pNfo.projectGroupGoPkg, string(filepath.Separator))

	pNfo.projectGroupName = pNfo.name
	if l := len(splitProjectGoPkg); l > 1 {
		switch len(splitProjectGoPkg) {
		case 1:
			pNfo.projectGroupName = splitProjectGoPkg[0]
		case 2:
			pNfo.projectGroupName = splitProjectGoPkg[1]
		default:
			pNfo.projectGroupName = splitProjectGoPkg[l-1]
		}
	}
	remplacer := strings.NewReplacer(".", "", "-", "")
	pNfo.projectGroupName = remplacer.Replace(pNfo.projectGroupName)

	return nil
}

func (pNfo *PkgNfo) setShortNameAndPrefix() {
	pNfo.prefix, pNfo.shortName = ExtractPrefix(pNfo.Name(), pNfo.DefaultPrefixes())
}

func (pNfo *PkgNfo) SetDefaultPrefixes(s string) error {
	pNfo.defaultPrefixes = NormalizeDefaultPrefixes(s)
	pNfo.setShortNameAndPrefix()

	return nil
}

func (pNfo PkgNfo) DefaultPrefixes() string   { return pNfo.defaultPrefixes }
func (pNfo PkgNfo) Prefix() string            { return pNfo.prefix }
func (pNfo PkgNfo) Name() string              { return pNfo.name }
func (pNfo PkgNfo) ShortName() string         { return pNfo.shortName }
func (pNfo PkgNfo) GoPkg() string             { return pNfo.goPkg }
func (pNfo PkgNfo) ProjectGroupGoPkg() string { return pNfo.projectGroupGoPkg }
func (pNfo PkgNfo) ProjectGroupName() string  { return pNfo.projectGroupName }
func (pNfo PkgNfo) Path() string              { return pNfo.path }

// Copied and re-worked from
// https://github.com/spf13/cobra/bl ob/master/cobra/cmd/helpers.go
func Path(inputPath string) (string, error) {
	// if no path is provided... assume CWD.
	if inputPath == "" {
		x, err := os.Getwd()
		if err != nil {
			return "", err
		}

		return x, nil
	}

	var projectPath string
	var projectBase string
	srcPath := SrcPath()

	// if provided, inspect for logical locations
	if strings.ContainsRune(inputPath, os.PathSeparator) {
		if filepath.IsAbs(inputPath) || filepath.HasPrefix(inputPath, string(os.PathSeparator)) {
			// if Absolute, use it
			projectPath = filepath.Clean(inputPath)
			return projectPath, nil
		}
		// If not absolute but contains slashes,
		// assuming it means create it from $GOPATH
		count := strings.Count(inputPath, string(os.PathSeparator))

		if count == 1 {
			projectPath = filepath.Join(srcPath, "github.com", inputPath)
		} else {
			projectPath = filepath.Join(srcPath, inputPath)
		}
		return projectPath, nil
	}

	// hardest case.. just a word.
	if projectBase == "" {
		x, err := os.Getwd()
		if err == nil {
			projectPath = filepath.Join(x, inputPath)
			return projectPath, nil
		}
		return "", err
	}

	projectPath = filepath.Join(srcPath, projectBase, inputPath)

	return projectPath, nil
}

func NormalizeDefaultPrefixes(s string) string {
	if s != "" {
		prefixes := strings.Split(GomeetDefaultPrefixes+","+s, ",")
		data := sort.StringSlice(prefixes)
		sort.Sort(data)
		n := set.Uniq(data)
		prefixes = data[:n]
		return strings.Join(prefixes, ",")
	}

	return GomeetDefaultPrefixes
}

func GomeetPkg() string {
	return strings.TrimSuffix(reflect.TypeOf(Empty{}).PkgPath(), "/utils/project/helpers")
}

func ParseCmd(s string) []string {
	r := regexp.MustCompile(`'.*?'|".*?"|\S+`)
	res := r.FindAllString(s, -1)
	for k, v := range res {
		mod := strings.Trim(v, " ")
		mod = strings.Trim(mod, "'")
		mod = strings.Trim(mod, `"`)
		mod = strings.Trim(mod, " ")

		res[k] = mod
	}
	return res
}

func Base(absPath string) string {
	rel, err := filepath.Rel(SrcPath(), absPath)
	if err != nil {
		return filepath.ToSlash(absPath)
	}
	return filepath.ToSlash(rel)
}

func ExtractPrefix(name, prefix string) (string, string) {
	prefix = NormalizeDefaultPrefixes(prefix)

	if prefix != "" {
		prefixes := strings.Split(prefix, ",")
		tv := false
		for _, v := range prefixes {
			v = strings.Trim(v, " ")
			if strings.HasPrefix(name, v) {
				name = strings.Replace(name, v, "", -1)
				prefix = v
				tv = true
				break
			}
		}
		if !tv {
			prefix = ""
		}
	}

	return prefix, name
}

func LastFromSplit(input, split string) string {
	rel := strings.Split(input, split)
	return rel[len(rel)-1]
}

func SrcPath() string {
	return filepath.Join(os.Getenv("GOPATH"), "src") + string(os.PathSeparator)
}
