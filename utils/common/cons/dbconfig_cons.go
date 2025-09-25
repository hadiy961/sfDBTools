package cons

// DeletionType represents the type of deletion operation
type DeletionType string

const (
	DeletionSingle   DeletionType = "single"
	DeletionMultiple DeletionType = "multiple"
	DeletionAll      DeletionType = "all"
)

// OperationType represents database config operations
type OperationType string

const (
	OperationShow     OperationType = "show"
	OperationValidate OperationType = "validate"
	OperationDelete   OperationType = "delete"
	OperationEdit     OperationType = "edit"
	OperationGenerate OperationType = "generate"
)
