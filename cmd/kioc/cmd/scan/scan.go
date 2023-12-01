package scan

import (
	"github.com/go-kid/ioc/cmd/kioc/creator"
	"github.com/go-kid/ioc/cmd/kioc/util"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var Scan = &cobra.Command{
	Use:   "scan",
	Short: "scan and register components",
	Run:   scan,
}

var (
	packageArg    string
	outputDirArg  string
	replacePkgArg []string
)

func init() {
	Scan.Flags().StringVarP(&packageArg, "package", "p", ".", "scan package path")
	Scan.Flags().StringVarP(&outputDirArg, "output_dir", "o", "./register", "register file path")
	Scan.Flags().StringArrayVarP(&replacePkgArg, "replace", "r", nil, "used when package not equal directory's name\n"+
		"exp:\n"+
		"directory name: /xxx.com/project/utils\n"+
		"package name: util\n"+
		"then: xxx.com/project/utils=>xxx.com/project/util")
}

func scan(cmd *cobra.Command, args []string) {
	var files []string
	_ = filepath.WalkDir(packageArg, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})

	c := util.GoCmd{}
	list, err := c.List()
	if err != nil {
		log.Fatal(err)
	}
	mod := list.Path

	var registers []*Register
	for _, path := range files {
		r, err := analyseFile(path)
		if err != nil {
			log.Fatal(err)
		}
		registers = append(registers, r...)
	}
	groups := lo.GroupBy(registers, func(item *Register) string {
		return item.Group
	})

	var replaceRules [][]string
	for _, expression := range replacePkgArg {
		replaceRules = append(replaceRules, strings.SplitN(expression, "=>", 2))
	}

	var creators []creator.FileCreator
	for group, registers := range groups {
		f := creator.NewGoFile("register", outputDirArg, "scan_"+group, false)
		imports := lo.Map(registers, func(item *Register, index int) string {
			importPath := filepath.Join(mod, item.Path)
			if len(replaceRules) > 0 {
				for _, rule := range replaceRules {
					importPath = strings.ReplaceAll(importPath, rule[0], rule[1])
				}
			}
			return importPath
		})
		imports = lo.Uniq(imports)
		f.SetAttribute("Imports", imports)
		f.SetAttribute("Components", registers)
		creators = append(creators, f)
	}

	err = creator.NewBatchCreator(creators...).Create()
	if err != nil {
		log.Fatal(err)
	}
}

type Register struct {
	Path  string
	Pkg   string
	Name  string
	Group string
	Kind  string
}

func analyseFile(path string) (registers []*Register, err error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	fset := token.NewFileSet()
	var f *ast.File
	f, err = parser.ParseFile(fset, "", string(bytes), parser.ParseComments)
	if err != nil {
		return
	}
	dir, _ := filepath.Split(path)
	registers, err = analyseToken(f)
	if err != nil {
		return
	}
	for _, r := range registers {
		r.Path = dir
	}
	return
}

func analyseToken(f *ast.File) (registers []*Register, err error) {
	//ast.Print(fset, f)

	registerCommentMatch, err := regexp.Compile("//\\s*\\S+\\s+@\\S+")
	if err != nil {
		return nil, err
	}

	ast.Inspect(f, func(node ast.Node) bool {
		if comment, ok := node.(*ast.Comment); ok &&
			registerCommentMatch.MatchString(comment.Text) {
			cmm := comment.Text
			cmm = strings.ReplaceAll(cmm, " ", "")
			arr := strings.SplitN(cmm, "@", 2)
			name, group := strings.TrimPrefix(arr[0], "//"), arr[1]
			var kind string
			ast.Inspect(f, func(node ast.Node) bool {
				if id, ok := node.(*ast.Ident); ok && id.Name == name {
					kind = id.Obj.Kind.String()
					return false
				}
				return true
			})
			registers = append(registers, &Register{
				Pkg:   f.Name.Name,
				Name:  name,
				Group: group,
				Kind:  kind,
			})
		}
		return true
	})
	return
}
