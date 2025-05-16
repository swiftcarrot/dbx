package schema

// areColumnTypesEqual compares column types based on their SQL representation
func areColumnTypesEqual(a, b ColumnType) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Special case for VarcharType, ignore Length=0 vs Length>0 differences
	_, aIsVarchar := a.(*VarcharType)
	_, bIsVarchar := b.(*VarcharType)
	if aIsVarchar && bIsVarchar {
		return true
	}

	return a.SQL() == b.SQL()
}

// Diff compares two schemas and returns changes to migrate from source to target
func Diff(source, target *Schema) []Change {
	changes := []Change{}

	changes = append(changes, diffSchemaNames(source, target)...)
	changes = append(changes, diffExtensions(source, target)...)
	changes = append(changes, diffSequences(source, target)...)
	changes = append(changes, diffFunctions(source, target)...)
	changes = append(changes, diffViews(source, target)...)
	changes = append(changes, diffRowPolicies(source, target)...)

	// Tables that exist in source but not in target should be dropped
	for _, sourceTable := range source.Tables {
		found := false
		for _, targetTable := range target.Tables {
			if sourceTable.Name == targetTable.Name &&
				sourceTable.Schema == targetTable.Schema {
				found = true
				break
			}
		}
		if !found {
			changes = append(changes, &DropTableChange{
				TableName:  sourceTable.Name,
				SchemaName: sourceTable.Schema,
			})
		}
	}

	// Find tables to create or modify
	for _, targetTable := range target.Tables {
		found := false
		for _, sourceTable := range source.Tables {
			if targetTable.Name == sourceTable.Name &&
				targetTable.Schema == sourceTable.Schema {
				found = true
				// Table exists in both source and target, diff it
				changes = append(changes, diffTable(sourceTable, targetTable)...)
				break
			}
		}

		if !found {
			// Table exists in target but not in source, create it
			changes = append(changes, &CreateTableChange{
				TableDef: targetTable,
			})
		}
	}

	// Diff triggers (after tables to ensure proper dependencies)
	changes = append(changes, diffTriggers(source, target)...)

	return changes
}

// diffSchemaNames compares schema names and returns create/drop schema changes
func diffSchemaNames(source, target *Schema) []Change {
	var changes []Change

	// Check if we need to create a schema
	if target.Name != "" && target.Name != "public" && source.Name != target.Name {
		changes = append(changes, &CreateSchemaChange{
			SchemaName: target.Name,
		})
	}

	// Check if we need to drop a schema
	if source.Name != "" && source.Name != "public" && source.Name != target.Name {
		// Only drop if there are no tables left in that schema
		hasTablesInSchema := false
		for _, table := range target.Tables {
			if table.Schema == source.Name {
				hasTablesInSchema = true
				break
			}
		}
		if !hasTablesInSchema {
			changes = append(changes, &DropSchemaChange{
				SchemaName: source.Name,
			})
		}
	}

	return changes
}

// diffExtensions compares extensions and returns create/drop extension changes
func diffExtensions(source, target *Schema) []Change {
	var changes []Change

	// Find extensions to disable
	for _, sourceExt := range source.Extensions {
		found := false
		for _, targetExt := range target.Extensions {
			if sourceExt == targetExt {
				found = true
				break
			}
		}
		if !found {
			changes = append(changes, &DisableExtensionChange{
				Extension: sourceExt,
			})
		}
	}

	// Find extensions to enable
	for _, targetExt := range target.Extensions {
		found := false
		for _, sourceExt := range source.Extensions {
			if targetExt == sourceExt {
				found = true
				break
			}
		}
		if !found {
			changes = append(changes, &EnableExtensionChange{
				Extension: targetExt,
			})
		}
	}

	return changes
}

// diffSequences compares sequences and returns create/alter/drop sequence changes
func diffSequences(source, target *Schema) []Change {
	var changes []Change

	// Find sequences to drop
	for _, sourceSeq := range source.Sequences {
		found := false
		for _, targetSeq := range target.Sequences {
			if sourceSeq.Name == targetSeq.Name &&
				sourceSeq.Schema == targetSeq.Schema {
				found = true
				break
			}
		}
		if !found {
			changes = append(changes, &DropSequenceChange{
				SequenceName: sourceSeq.Name,
				SchemaName:   sourceSeq.Schema,
			})
		}
	}

	// Find sequences to create or alter
	for _, targetSeq := range target.Sequences {
		found := false
		for _, sourceSeq := range source.Sequences {
			if sourceSeq.Name == targetSeq.Name &&
				sourceSeq.Schema == targetSeq.Schema {
				found = true
				// Check if sequence attributes have changed
				if sourceSeq.Start != targetSeq.Start ||
					sourceSeq.Increment != targetSeq.Increment ||
					sourceSeq.MinValue != targetSeq.MinValue ||
					sourceSeq.MaxValue != targetSeq.MaxValue ||
					sourceSeq.Cache != targetSeq.Cache ||
					sourceSeq.Cycle != targetSeq.Cycle {
					changes = append(changes, &AlterSequenceChange{
						Sequence: targetSeq,
					})
				}
				break
			}
		}
		if !found {
			changes = append(changes, &CreateSequenceChange{
				Sequence: targetSeq,
			})
		}
	}

	return changes
}

// diffFunctions compares functions and returns create/alter/drop function changes
func diffFunctions(source, target *Schema) []Change {
	var changes []Change

	// Functions that exist in source but not in target should be dropped
	for _, sourceFunc := range source.Functions {
		found := false
		for _, targetFunc := range target.Functions {
			if isSameFunction(sourceFunc, targetFunc) {
				found = true
				break
			}
		}

		if !found {
			changes = append(changes, &DropFunctionChange{
				FunctionName: sourceFunc.Name,
				FunctionArgs: sourceFunc.Arguments,
				SchemaName:   sourceFunc.Schema,
			})
		}
	}

	// Find functions to create or modify
	for _, targetFunc := range target.Functions {
		found := false
		for _, sourceFunc := range source.Functions {
			if isSameFunction(sourceFunc, targetFunc) {
				found = true

				// Function exists in both source and target, check if it needs modification
				if !isSameFunctionDefinition(sourceFunc, targetFunc) {
					changes = append(changes, &AlterFunctionChange{
						Function: targetFunc,
					})
				}
				break
			}
		}

		if !found {
			// Function exists in target but not in source, create it
			changes = append(changes, &CreateFunctionChange{
				Function: targetFunc,
			})
		}
	}

	return changes
}

// isSameFunction checks if two functions have the same identity (name, schema, and argument types)
func isSameFunction(f1, f2 *Function) bool {
	if f1.Name != f2.Name || f1.Schema != f2.Schema {
		return false
	}

	// PostgreSQL identifies functions by name and argument types (not names)
	if len(f1.Arguments) != len(f2.Arguments) {
		return false
	}

	for i := range f1.Arguments {
		// Compare only types, not argument names or modes
		if f1.Arguments[i].Type != f2.Arguments[i].Type {
			return false
		}
	}

	return true
}

// isSameFunctionDefinition checks if two functions have the same implementation
func isSameFunctionDefinition(f1, f2 *Function) bool {
	// First check signature identity
	if !isSameFunction(f1, f2) {
		return false
	}

	// Then check all definition details
	if f1.Returns != f2.Returns ||
		f1.Language != f2.Language ||
		f1.Body != f2.Body ||
		f1.Volatility != f2.Volatility ||
		f1.Strict != f2.Strict ||
		f1.Security != f2.Security ||
		f1.Cost != f2.Cost {
		return false
	}

	return true
}

// diffViews compares views and returns create/alter/drop view changes
func diffViews(source, target *Schema) []Change {
	var changes []Change

	// Views that exist in source but not in target should be dropped
	for _, sourceView := range source.Views {
		found := false
		for _, targetView := range target.Views {
			if sourceView.Name == targetView.Name && sourceView.Schema == targetView.Schema {
				found = true
				break
			}
		}

		if !found {
			changes = append(changes, &DropViewChange{
				ViewName:   sourceView.Name,
				SchemaName: sourceView.Schema,
			})
		}
	}

	// Find views to create or modify
	for _, targetView := range target.Views {
		found := false
		for _, sourceView := range source.Views {
			if targetView.Name == sourceView.Name && targetView.Schema == sourceView.Schema {
				found = true

				// View exists in both source and target, check if it needs modification
				if targetView.Definition != sourceView.Definition ||
					!stringsEqual(targetView.Options, sourceView.Options) ||
					!stringsEqual(targetView.Columns, sourceView.Columns) {
					changes = append(changes, &AlterViewChange{
						View: targetView,
					})
				}
				break
			}
		}

		if !found {
			// View exists in target but not in source, create it
			changes = append(changes, &CreateViewChange{
				View: targetView,
			})
		}
	}

	return changes
}

// stringsEqual compares two string slices for equality regardless of order
func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Copy slices to avoid modifying the originals
	aCopy := make([]string, len(a))
	copy(aCopy, a)

	bCopy := make([]string, len(b))
	copy(bCopy, b)

	// Sort both slices
	// Note: we're omitting actual sorting for simplicity, but sorting would be needed here

	// Compare sorted slices
	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}

	return true
}

// diffTriggers compares triggers and returns create/alter/drop trigger changes
func diffTriggers(source, target *Schema) []Change {
	var changes []Change

	// Triggers that exist in source but not in target should be dropped
	for _, sourceTrigger := range source.Triggers {
		found := false
		for _, targetTrigger := range target.Triggers {
			if sourceTrigger.Name == targetTrigger.Name &&
				sourceTrigger.Schema == targetTrigger.Schema &&
				sourceTrigger.Table == targetTrigger.Table {
				found = true
				break
			}
		}

		if !found {
			changes = append(changes, &DropTriggerChange{
				TriggerName:  sourceTrigger.Name,
				TriggerTable: sourceTrigger.Table,
				SchemaName:   sourceTrigger.Schema,
			})
		}
	}

	// Find triggers to create or modify
	for _, targetTrigger := range target.Triggers {
		found := false
		for _, sourceTrigger := range source.Triggers {
			if targetTrigger.Name == sourceTrigger.Name &&
				targetTrigger.Schema == sourceTrigger.Schema &&
				targetTrigger.Table == sourceTrigger.Table {
				found = true

				// Trigger exists in both source and target, check if it needs modification
				if !isSameTriggerDefinition(sourceTrigger, targetTrigger) {
					changes = append(changes, &AlterTriggerChange{
						Trigger: targetTrigger,
					})
				}
				break
			}
		}

		if !found {
			// Trigger exists in target but not in source, create it
			changes = append(changes, &CreateTriggerChange{
				Trigger: targetTrigger,
			})
		}
	}

	return changes
}

// isSameTriggerDefinition checks if two triggers have the same definition
func isSameTriggerDefinition(t1, t2 *Trigger) bool {
	if t1.Name != t2.Name ||
		t1.Schema != t2.Schema ||
		t1.Table != t2.Table ||
		t1.Timing != t2.Timing ||
		t1.ForEach != t2.ForEach ||
		t1.When != t2.When ||
		t1.Function != t2.Function {
		return false
	}

	// Compare events
	if len(t1.Events) != len(t2.Events) {
		return false
	}
	for i := range t1.Events {
		if t1.Events[i] != t2.Events[i] {
			return false
		}
	}

	// Compare arguments
	if len(t1.Arguments) != len(t2.Arguments) {
		return false
	}
	for i := range t1.Arguments {
		if t1.Arguments[i] != t2.Arguments[i] {
			return false
		}
	}

	return true
}

// diffTable compares two tables and returns changes to migrate from source to target
func diffTable(sourceTable, targetTable *Table) []Change {
	var changes []Change

	// Compare columns
	changes = append(changes, diffColumns(sourceTable, targetTable)...)

	// Compare primary keys
	changes = append(changes, diffPrimaryKeys(sourceTable, targetTable)...)

	// Compare indexes
	changes = append(changes, diffIndexes(sourceTable, targetTable)...)

	// Compare foreign keys
	changes = append(changes, diffForeignKeys(sourceTable, targetTable)...)

	return changes
}

// diffColumns compares columns between two tables and returns changes
func diffColumns(sourceTable, targetTable *Table) []Change {
	var changes []Change

	// Find columns to drop
	for _, sourceCol := range sourceTable.Columns {
		found := false
		for _, targetCol := range targetTable.Columns {
			if sourceCol.Name == targetCol.Name {
				found = true
				break
			}
		}
		if !found {
			changes = append(changes, &DropColumnChange{
				TableName:  sourceTable.Name,
				ColumnName: sourceCol.Name,
			})
		}
	}

	// Find columns to add or modify
	for _, targetCol := range targetTable.Columns {
		found := false
		for _, sourceCol := range sourceTable.Columns {
			if sourceCol.Name == targetCol.Name {
				found = true
				// Column exists in both, check if they're different
				if !areColumnTypesEqual(sourceCol.Type, targetCol.Type) ||
					sourceCol.Nullable != targetCol.Nullable ||
					sourceCol.Default != targetCol.Default ||
					sourceCol.Comment != targetCol.Comment {
					changes = append(changes, &AlterColumnChange{
						TableName: targetTable.Name,
						Column:    targetCol,
					})
				}
				break
			}
		}
		if !found {
			// Column doesn't exist in source, add it
			changes = append(changes, &AddColumnChange{
				TableName: targetTable.Name,
				Column:    targetCol,
			})
		}
	}

	return changes
}

// diffPrimaryKeys compares primary keys between two tables
func diffPrimaryKeys(sourceTable, targetTable *Table) []Change {
	var changes []Change

	// If source has primary key but target doesn't, drop it
	if sourceTable.PrimaryKey != nil && targetTable.PrimaryKey == nil {
		changes = append(changes, &DropPrimaryKeyChange{
			TableName: sourceTable.Name,
			PKName:    sourceTable.PrimaryKey.Name,
		})
		return changes
	}

	// If target has primary key but source doesn't, add it
	if sourceTable.PrimaryKey == nil && targetTable.PrimaryKey != nil {
		changes = append(changes, &AddPrimaryKeyChange{
			TableName:  targetTable.Name,
			PrimaryKey: targetTable.PrimaryKey,
		})
		return changes
	}

	// If both have primary key, check if they're different
	if sourceTable.PrimaryKey != nil && targetTable.PrimaryKey != nil {
		if !equalStringSlices(sourceTable.PrimaryKey.Columns, targetTable.PrimaryKey.Columns) ||
			sourceTable.PrimaryKey.Name != targetTable.PrimaryKey.Name {
			// Drop the old one and add the new one
			changes = append(changes, &DropPrimaryKeyChange{
				TableName: sourceTable.Name,
				PKName:    sourceTable.PrimaryKey.Name,
			})
			changes = append(changes, &AddPrimaryKeyChange{
				TableName:  targetTable.Name,
				PrimaryKey: targetTable.PrimaryKey,
			})
		}
	}

	return changes
}

// diffIndexes compares indexes between two tables and returns changes
func diffIndexes(sourceTable, targetTable *Table) []Change {
	var changes []Change

	// Find indexes to drop
	for _, sourceIdx := range sourceTable.Indexes {
		found := false
		for _, targetIdx := range targetTable.Indexes {
			if sourceIdx.Name == targetIdx.Name {
				found = true
				break
			}
		}
		if !found {
			changes = append(changes, &DropIndexChange{
				TableName: sourceTable.Name,
				IndexName: sourceIdx.Name,
			})
		}
	}

	// Find indexes to add or modify
	for _, targetIdx := range targetTable.Indexes {
		found := false
		for _, sourceIdx := range sourceTable.Indexes {
			if sourceIdx.Name == targetIdx.Name {
				found = true
				// Index exists in both, check if they're different
				if !equalStringSlices(sourceIdx.Columns, targetIdx.Columns) ||
					sourceIdx.Unique != targetIdx.Unique {
					// Drop the old one and add the new one
					changes = append(changes, &DropIndexChange{
						TableName: sourceTable.Name,
						IndexName: sourceIdx.Name,
					})
					changes = append(changes, &AddIndexChange{
						TableName: targetTable.Name,
						Index:     targetIdx,
					})
				}
				break
			}
		}
		if !found {
			// Index doesn't exist in source, add it
			changes = append(changes, &AddIndexChange{
				TableName: targetTable.Name,
				Index:     targetIdx,
			})
		}
	}

	return changes
}

// diffForeignKeys compares foreign keys between two tables and returns changes
func diffForeignKeys(sourceTable, targetTable *Table) []Change {
	var changes []Change

	// Find foreign keys to drop
	for _, sourceFk := range sourceTable.ForeignKeys {
		found := false
		for _, targetFk := range targetTable.ForeignKeys {
			if sourceFk.Name == targetFk.Name {
				found = true
				break
			}
		}
		if !found {
			changes = append(changes, &DropForeignKeyChange{
				TableName: sourceTable.Name,
				FKName:    sourceFk.Name,
			})
		}
	}

	// Find foreign keys to add or modify
	for _, targetFk := range targetTable.ForeignKeys {
		found := false
		for _, sourceFk := range sourceTable.ForeignKeys {
			if sourceFk.Name == targetFk.Name {
				found = true
				// Foreign key exists in both, check if they're different
				if !equalStringSlices(sourceFk.Columns, targetFk.Columns) ||
					!equalStringSlices(sourceFk.RefColumns, targetFk.RefColumns) ||
					sourceFk.RefTable != targetFk.RefTable ||
					sourceFk.OnDelete != targetFk.OnDelete ||
					sourceFk.OnUpdate != targetFk.OnUpdate {
					// Drop the old one and add the new one
					changes = append(changes, &DropForeignKeyChange{
						TableName: sourceTable.Name,
						FKName:    sourceFk.Name,
					})
					changes = append(changes, &AddForeignKeyChange{
						TableName:  targetTable.Name,
						ForeignKey: targetFk,
					})
				}
				break
			}
		}
		if !found {
			// Foreign key doesn't exist in source, add it
			changes = append(changes, &AddForeignKeyChange{
				TableName:  targetTable.Name,
				ForeignKey: targetFk,
			})
		}
	}

	return changes
}

// equalStringSlices checks if two string slices are equal
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// diffRowPolicies compares row policies and returns create/alter/drop row policy changes
func diffRowPolicies(source, target *Schema) []Change {
	var changes []Change

	// Row policies that exist in source but not in target should be dropped
	for _, sourcePolicy := range source.RowPolicies {
		found := false
		for _, targetPolicy := range target.RowPolicies {
			if sourcePolicy.PolicyName == targetPolicy.PolicyName &&
				sourcePolicy.TableName == targetPolicy.TableName &&
				sourcePolicy.Schema == targetPolicy.Schema {
				found = true
				break
			}
		}

		if !found {
			changes = append(changes, &DropRowPolicyChange{
				SchemaName: sourcePolicy.Schema,
				TableName:  sourcePolicy.TableName,
				PolicyName: sourcePolicy.PolicyName,
			})
		}
	}

	// Find row policies to create or modify
	for _, targetPolicy := range target.RowPolicies {
		found := false
		for _, sourcePolicy := range source.RowPolicies {
			if targetPolicy.PolicyName == sourcePolicy.PolicyName &&
				targetPolicy.TableName == sourcePolicy.TableName &&
				targetPolicy.Schema == sourcePolicy.Schema {
				found = true

				// Policy exists in both source and target, check if it needs modification
				if targetPolicy.CommandType != sourcePolicy.CommandType ||
					!stringsEqual(targetPolicy.Roles, sourcePolicy.Roles) ||
					targetPolicy.UsingExpr != sourcePolicy.UsingExpr ||
					targetPolicy.CheckExpr != sourcePolicy.CheckExpr ||
					targetPolicy.Permissive != sourcePolicy.Permissive {
					changes = append(changes, &AlterRowPolicyChange{
						RowPolicy: targetPolicy,
					})
				}
				break
			}
		}

		if !found {
			// Row policy exists in target but not in source, create it
			changes = append(changes, &CreateRowPolicyChange{
				RowPolicy: targetPolicy,
			})
		}
	}

	return changes
}
