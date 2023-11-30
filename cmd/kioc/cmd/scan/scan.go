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
	packageArg   string
	outputDirArg string
)

func init() {
	Scan.Flags().StringVarP(&packageArg, "package", "p", ".", "scan package path")
	Scan.Flags().StringVarP(&outputDirArg, "output_dir", "o", "./register", "register file path")
}

func scan(cmd *cobra.Command, args []string) {
	var files []string
	_ = filepath.WalkDir(packageArg, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})

	var registers []*Register
	for _, file := range files {
		fileBytes, err := os.ReadFile(file)
		if err != nil {
			return
		}
		r, err := analyseFile(fileBytes)
		if err != nil {
			log.Fatal(err)
		}
		registers = append(registers, r...)
	}
	groups := lo.GroupBy(registers, func(item *Register) string {
		return item.Group
	})

	var creators []creator.FileCreator
	for group, registers := range groups {
		f := creator.NewGoFile("register", outputDirArg, "scan_"+group, false)
		paths := lo.Map(registers, func(item *Register, index int) string {
			return item.Path
		})
		paths = lo.Uniq(paths)
		f.SetAttribute("Imports", paths)
		f.SetAttribute("Components", registers)
		creators = append(creators, f)
	}

	err := creator.NewBatchCreator(creators...).Create()
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

func analyseFile(bytes []byte) (registers []*Register, err error) {
	fset := token.NewFileSet()
	var f *ast.File
	f, err = parser.ParseFile(fset, "", string(bytes), parser.ParseComments)
	if err != nil {
		return
	}
	//ast.Print(fset, f)

	c := util.GoCmd{}
	list, err := c.List()
	if err != nil {
		return nil, err
	}
	mod := list.Path

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
				Path:  filepath.Join(mod, f.Name.Name),
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
