package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/moroz/uuidv7-go"
	"time"
)

// UploadTask holds the schema definition for the UploadTask entity.
type UploadTask struct {
	ent.Schema
}

// Fields of the UploadTask.
func (UploadTask) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuidv7.UUID{}).Default(uuidv7.Generate).Immutable().Unique(),
		field.Enum("type").Values(string(MediaTypePhoto), string(MediaTypeVideo), string(MediaTypeAnimation)),
		field.Bytes("data").Optional(),
		field.Bool("is_processed").Default(false),
		field.Time("created_at").Default(time.Now),
		field.Time("sent_at").Optional(),
	}
}

// Edges of the UploadTask.
func (UploadTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("image_hash", ImageHash.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)).
			Unique(),
	}
}

func (UploadTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_processed"),
		index.Fields("type"),
	}
}
