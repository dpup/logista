// PROPOSED IMPLEMENTATION for formatter.go

// Add this function to the TemplateFormatter struct

// getFunc retrieves a field by name from a data map
// This provides syntactic sugar to access fields with dots in their names
// Instead of writing {{index . "grpc.service"}}, you can write {{get "grpc.service" .}}
func (f *TemplateFormatter) getFunc(name string, data map[string]interface{}) interface{} {
    return data[name]
}

// Then add this to the NewTemplateFormatter function:
tmpl := template.New("formatter").Funcs(template.FuncMap{
    // Value formatting
    "date": formatter.dateFunc,
    "pad":  formatter.padFunc,
    
    // Color functions
    "color":        formatter.colorFunc,
    "colorByLevel": formatter.colorByLevelFunc,
    "bold":         formatter.boldFunc,
    "italic":       formatter.italicFunc,
    "underline":    formatter.underlineFunc,
    "dim":          formatter.dimFunc,
    
    // Field filtering and categorization
    "isStandardField":     formatter.isStandardFieldFunc,
    "hasPrefix":           formatter.hasPrefixFunc,
    "getFields":           formatter.getFieldsFunc,
    "getFieldsWithout":    formatter.getFieldsWithoutFunc,
    "getFieldsWithPrefix": formatter.getFieldsWithPrefixFunc,
    
    // New field access helper
    "get": formatter.getFunc,
})