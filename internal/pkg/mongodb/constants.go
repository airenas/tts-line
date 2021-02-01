package mongodb

const (
	textTable = "text"
)

var indexData = []IndexData{
	newIndexData(textTable, []string{"id", "type"}, false)}
