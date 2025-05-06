package schema

import (
	"bytes"
	"fmt"
	"strings"
)

// String implements the fmt.Stringer interface, returning a string representation of the schema
func (s *Schema) String() string {
	var buf bytes.Buffer
	buf.WriteString("Schema:")

	// Print schema name if set
	if s.Name != "" {
		fmt.Fprintf(&buf, "\n  Schema Name: %s", s.Name)
	}

	// Print extensions if any
	if len(s.Extensions) > 0 {
		buf.WriteString("\n  Extensions:")
		for _, ext := range s.Extensions {
			fmt.Fprintf(&buf, "\n    %s", ext)
		}
		buf.WriteString("\n")
	}

	// Print sequences if any
	if len(s.Sequences) > 0 {
		buf.WriteString("\n  Sequences:")
		for _, seq := range s.Sequences {
			schemaPrefix := ""
			if seq.Schema != "" {
				schemaPrefix = seq.Schema + "."
			}

			fmt.Fprintf(&buf, "\n    %s%s (START %d", schemaPrefix, seq.Name, seq.Start)
			if seq.Increment != 1 {
				fmt.Fprintf(&buf, ", INCREMENT %d", seq.Increment)
			}
			if seq.MinValue != 1 {
				fmt.Fprintf(&buf, ", MINVALUE %d", seq.MinValue)
			}
			if seq.MaxValue != 9223372036854775807 { // Default max for bigint
				fmt.Fprintf(&buf, ", MAXVALUE %d", seq.MaxValue)
			}
			if seq.Cache != 1 {
				fmt.Fprintf(&buf, ", CACHE %d", seq.Cache)
			}
			if seq.Cycle {
				fmt.Fprintf(&buf, ", CYCLE")
			}
			fmt.Fprintf(&buf, ")")
		}
		buf.WriteString("\n")
	}

	// Print functions if any
	if len(s.Functions) > 0 {
		buf.WriteString("\n  Functions:")
		for _, fn := range s.Functions {
			schemaPrefix := ""
			if fn.Schema != "" {
				schemaPrefix = fn.Schema + "."
			}

			// Format function arguments
			args := make([]string, len(fn.Arguments))
			for i, arg := range fn.Arguments {
				argStr := ""
				if arg.Name != "" {
					argStr += arg.Name + " "
				}

				if arg.Mode != "IN" { // IN is default, so only show if different
					argStr += arg.Mode + " "
				}

				argStr += arg.Type

				if arg.Default != "" {
					argStr += fmt.Sprintf(" DEFAULT %s", arg.Default)
				}

				args[i] = argStr
			}

			fmt.Fprintf(&buf, "\n    %s%s(%s) RETURNS %s",
				schemaPrefix, fn.Name, strings.Join(args, ", "), fn.Returns)

			fmt.Fprintf(&buf, "\n      LANGUAGE %s", fn.Language)

			if fn.Volatility != "VOLATILE" { // Only print if not default
				fmt.Fprintf(&buf, " %s", fn.Volatility)
			}

			if fn.Strict {
				fmt.Fprintf(&buf, " STRICT")
			}

			if fn.Security != "INVOKER" { // Only print if not default
				fmt.Fprintf(&buf, " SECURITY %s", fn.Security)
			}

			fmt.Fprintf(&buf, " COST %d", fn.Cost)

			// Print a condensed version of the body (first 50 chars)
			bodyPreview := fn.Body
			if len(bodyPreview) > 50 {
				bodyPreview = bodyPreview[:50] + "..."
			}
			fmt.Fprintf(&buf, "\n      Body: %s", strings.ReplaceAll(bodyPreview, "\n", " "))
		}
		buf.WriteString("\n")
	}

	// Print views if any
	if len(s.Views) > 0 {
		buf.WriteString("\n  Views:")
		for _, view := range s.Views {
			schemaPrefix := ""
			if view.Schema != "" {
				schemaPrefix = view.Schema + "."
			}

			fmt.Fprintf(&buf, "\n    %s%s", schemaPrefix, view.Name)

			if len(view.Columns) > 0 {
				fmt.Fprintf(&buf, " (%s)", strings.Join(view.Columns, ", "))
			}

			if len(view.Options) > 0 {
				fmt.Fprintf(&buf, " WITH (%s)", strings.Join(view.Options, ", "))
			}

			// Print a condensed version of the definition (first 50 chars)
			defPreview := view.Definition
			if len(defPreview) > 50 {
				defPreview = defPreview[:50] + "..."
			}
			fmt.Fprintf(&buf, "\n      Definition: %s", strings.ReplaceAll(defPreview, "\n", " "))
		}
		buf.WriteString("\n")
	}

	// Print triggers if any
	if len(s.Triggers) > 0 {
		buf.WriteString("\n  Triggers:")
		for _, trigger := range s.Triggers {
			schemaPrefix := ""
			if trigger.Schema != "" {
				schemaPrefix = trigger.Schema + "."
			}

			fmt.Fprintf(&buf, "\n    %s%s ON %s",
				schemaPrefix, trigger.Name, trigger.Table)

			fmt.Fprintf(&buf, "\n      %s %s FOR EACH %s",
				trigger.Timing,
				strings.Join(trigger.Events, " OR "),
				trigger.ForEach)

			if trigger.When != "" {
				fmt.Fprintf(&buf, " WHEN (%s)", trigger.When)
			}

			fmt.Fprintf(&buf, " EXECUTE FUNCTION %s", trigger.Function)

			if len(trigger.Arguments) > 0 {
				fmt.Fprintf(&buf, "(%s)", strings.Join(trigger.Arguments, ", "))
			}
		}
		buf.WriteString("\n")
	}

	if len(s.Tables) == 0 {
		buf.WriteString("\n  No tables defined")
		return buf.String()
	}

	for _, table := range s.Tables {
		fmt.Fprintf(&buf, "\n  Table: %s\n", table.Name)

		// Print columns
		buf.WriteString("    Columns:")
		for _, col := range table.Columns {
			nullableStr := "NOT NULL"
			if col.Nullable {
				nullableStr = "NULL"
			}

			defaultStr := ""
			if col.Default != "" {
				defaultStr = fmt.Sprintf(" DEFAULT %s", col.Default)
			}

			commentStr := ""
			if col.Comment != "" {
				commentStr = fmt.Sprintf(" -- %s", col.Comment)
			}

			fmt.Fprintf(&buf, "\n      %s %s %s%s%s", col.Name, col.Type, nullableStr, defaultStr, commentStr)
		}

		// Print primary key if exists
		if table.PrimaryKey != nil {
			fmt.Fprintf(&buf, "\n    Primary Key: %s (%s)",
				table.PrimaryKey.Name,
				strings.Join(table.PrimaryKey.Columns, ", "))
		}

		// Print indexes
		if len(table.Indexes) > 0 {
			buf.WriteString("\n    Indexes:")
			for _, idx := range table.Indexes {
				uniqueStr := ""
				if idx.Unique {
					uniqueStr = "UNIQUE "
				}
				fmt.Fprintf(&buf, "\n      %s%s (%s)", uniqueStr, idx.Name, strings.Join(idx.Columns, ", "))
			}
		}

		// Print foreign keys
		if len(table.ForeignKeys) > 0 {
			buf.WriteString("\n    Foreign Keys:")
			for _, fk := range table.ForeignKeys {
				onDeleteStr := ""
				if fk.OnDelete != "" {
					onDeleteStr = fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
				}

				onUpdateStr := ""
				if fk.OnUpdate != "" {
					onUpdateStr = fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
				}

				fmt.Fprintf(&buf, "\n      %s: (%s) REFERENCES %s (%s)%s%s",
					fk.Name,
					strings.Join(fk.Columns, ", "),
					fk.RefTable,
					strings.Join(fk.RefColumns, ", "),
					onDeleteStr,
					onUpdateStr)
			}
		}

		buf.WriteString("\n") // Add a blank line between tables
	}

	return buf.String()
}
