package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type ImageHash struct {
	ent.Schema
}

func (ImageHash) Fields() []ent.Field {
	return []ent.Field{
		field.String("image_hash").NotEmpty(),
	}
}

func (ImageHash) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).Ref("image_hash").Unique(),
		edge.From("upload_task", UploadTask.Type).Ref("image_hash").Unique(),
	}
}
