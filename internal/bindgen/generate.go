package bindgen

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func GenerateBindings(headerPath string, libName string) (GenerateResult, error) {
	cmd := exec.Command("clang", "-Xclang", "-ast-dump=json", "-fsyntax-only", headerPath)
	out, err := cmd.Output()
	if err != nil {
		return GenerateResult{}, fmt.Errorf("clang error: %v (output: %s)", err, string(out))
	}

	var root ClangASTNode
	if err := json.Unmarshal(out, &root); err != nil {
		return GenerateResult{}, fmt.Errorf("json unmarshal error: %v", err)
	}

	structs := make(map[string]string)
	structOrder := []string{}
	var externs []string
	skipped := 0

	enums := []string{}
	seenStructs := make(map[string]bool)

	for _, node := range root.Inner {
		if node.Kind == "RecordDecl" && node.CompleteDefinition && node.Name != "" {
			if seenStructs[node.Name] {
				continue
			}
			body, ok := parseStructFields(node)
			if ok {
				structs[node.Name] = body
				structOrder = append(structOrder, node.Name)
				seenStructs[node.Name] = true
			}
		} else if node.Kind == "EnumDecl" && node.Name != "" {
			enumBody := parseEnum(node)
			if enumBody != "" {
				enums = append(enums, enumBody)
			}
		} else if node.Kind == "FunctionDecl" && node.Name != "" {
			decl, ok := parseFunction(node)
			if ok {
				externs = append(externs, decl)
			} else {
				skipped++
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("' AUTO-GENERATED bindings for %s\n", libName))
	sb.WriteString(fmt.Sprintf("# link \"%s\"\n\n", libName))

	for _, e := range enums {
		sb.WriteString(e + "\n\n")
	}

	for _, name := range structOrder {
		sb.WriteString(fmt.Sprintf("pub struct %s {\n%s\n}\n\n", name, structs[name]))
	}

	for _, ext := range externs {
		sb.WriteString(ext + "\n\n")
	}

	return GenerateResult{
		Content: sb.String(),
		Structs: len(structs),
		Externs: len(externs),
		Skipped: skipped,
	}, nil
}

func parseEnum(node ClangASTNode) string {
	var variants []string
	for _, child := range node.Inner {
		if child.Kind == "EnumConstantDecl" {
			variants = append(variants, child.Name)
		}
	}
	if len(variants) == 0 {
		return ""
	}
	return fmt.Sprintf("pub enum %s { %s }", node.Name, strings.Join(variants, ", "))
}

func parseStructFields(node ClangASTNode) (string, bool) {
	var fields []string
	for _, child := range node.Inner {
		if child.Kind == "FieldDecl" {
			kodaeType, ok := mapCTypeToKodae(child.Type.QualType)
			if !ok {
				return "", false // Skip structs with unmappable fields
			}
			fields = append(fields, fmt.Sprintf("  %s: %s", child.Name, kodaeType))
		}
	}
	if len(fields) == 0 {
		return "", false
	}
	return strings.Join(fields, "\n"), true
}

func parseFunction(node ClangASTNode) (string, bool) {
	// Function type is in QualType, e.g., "void (int, int, float, struct Color)"
	// But it's easier to look at ParmVarDecl children
	retType, ok := mapCTypeToKodae(strings.Split(node.Type.QualType, " (")[0])
	if !ok {
		return "", false
	}

	var params []string
	for _, child := range node.Inner {
		if child.Kind == "ParmVarDecl" {
			pType, ok := mapCTypeToKodae(child.Type.QualType)
			if !ok {
				return "", false
			}
			params = append(params, fmt.Sprintf("%s: %s", child.Name, pType))
		}
	}

	return fmt.Sprintf("extern fn %s(%s) -> %s", node.Name, strings.Join(params, ", "), retType), true
}

func mapCTypeToKodae(cType string) (string, bool) {
	cType = strings.TrimSpace(cType)
	cType = strings.TrimPrefix(cType, "const ")
	
	// Handle pointers
	if strings.HasSuffix(cType, "*") {
		inner := strings.TrimSpace(strings.TrimSuffix(cType, "*"))
		if inner == "char" || inner == "void" || strings.HasPrefix(inner, "unsigned char") {
			return "ptr[byte]", true
		}
		// Generic pointer for now, maybe more specific later
		return "ptr[byte]", true 
	}

	// Handle "struct Name"
	if strings.HasPrefix(cType, "struct ") {
		return strings.TrimPrefix(cType, "struct "), true
	}

	// Handle "enum Name"
	if strings.HasPrefix(cType, "enum ") {
		return strings.TrimPrefix(cType, "enum "), true
	}

	switch cType {
	case "int", "signed int", "long", "long long":
		return "int", true
	case "unsigned int", "unsigned long", "unsigned long long":
		return "int", true // Kodae's int is 64-bit
	case "float":
		return "f32", true
	case "double":
		return "float", true
	case "bool", "_Bool":
		return "bool", true
	case "unsigned char", "uint8_t":
		return "u8", true
	case "char", "signed char", "int8_t":
		return "u8", true // Or i8 if Kodae has it
	case "short", "int16_t":
		return "int", true
	case "unsigned short", "uint16_t":
		return "int", true
	case "void":
		return "void", true
	}

	return "", false
}
