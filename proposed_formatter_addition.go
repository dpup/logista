// This is a proposed addition to formatter.go

// fieldFunc creates a function that retrieves values from the current data context
// It makes templates cleaner when accessing fields with dots in their names
func (f *TemplateFormatter) fieldFunc(data map[string]interface{}, name string) interface{} {
    // Check if the field name contains a dot
    if strings.Contains(name, ".") {
        // Access using the field name directly
        return data[name]
    }
    // For regular fields, access normally
    return data[name]
}

// Alternative implementation #1 - Using a curry function approach
// This creates a function that takes the data and returns another function that takes the field name
func (f *TemplateFormatter) withFieldFunc() interface{} {
    return func(data map[string]interface{}) interface{} {
        return func(name string) interface{} {
            if strings.Contains(name, ".") {
                return data[name]
            }
            return data[name]
        }
    }
}

// Alternative implementation #2 - Using dot access with index
// This shows how to register the template functions in NewTemplateFormatter:
tmpl := template.New("formatter").Funcs(template.FuncMap{
    // ... existing functions
    
    // Simple field getter that needs both data and field name
    "field": formatter.fieldFunc,
    
    // Function that allows you to write {{with $ := .}}{{field "name" $}}{{end}}
    "withField": formatter.withFieldFunc,
})