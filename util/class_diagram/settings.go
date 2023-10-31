package class_diagram

import "strconv"

func NamespaceSeparator(s string) string {
	return "set namespaceSeparator " + s
}

func GroupInheritance(s int) string {
	return "skinparam groupInheritance " + strconv.Itoa(s)
}
