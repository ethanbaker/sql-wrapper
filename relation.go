package sql_wrapper

// RelationType represents different foreign key relationships
type RelationType string

const (
	UndefinedRelationType RelationType = ""
	OneToOne              RelationType = "one-to-one"
	OneToMany             RelationType = "one-to-many"
	ManyToOne             RelationType = "many-to-one"
	ManyToMany            RelationType = "many-to-many"
)
