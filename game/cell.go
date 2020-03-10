package game

// Cell struct is represents the military and terrain state of a cell
type Cell struct {
	Armies  int
	Type    CellType
	Faction int
}

// CellType is a generic type for the terrain type of a cell
type CellType int

const (
	// Plain indicates an Open Plains cell
	Plain CellType = iota
	// City cell
	City
	// General cell
	General
)
