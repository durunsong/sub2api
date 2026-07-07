package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// IPBan holds globally banned client IP/CIDR rules.
type IPBan struct {
	ent.Schema
}

func (IPBan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ip_bans"},
	}
}

func (IPBan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (IPBan) Fields() []ent.Field {
	return []ent.Field{
		field.String("pattern").MaxLen(64).NotEmpty().Comment("Banned IP or CIDR pattern"),
		field.String("status").MaxLen(20).Default("active").Comment("active/inactive"),
		field.String("reason").MaxLen(255).Optional().Nillable().Comment("Ban reason"),
		field.String("source").MaxLen(50).Default("manual").Comment("Ban source"),
		field.Int64("created_by").Optional().Nillable().Comment("Admin user id that created the ban"),
		field.Time("expires_at").Optional().Nillable().Comment("Expiration time; null means permanent"),
		field.Time("last_hit_at").Optional().Nillable().Comment("Last time this rule blocked a request"),
		field.Int64("hit_count").Default(0).Comment("Number of blocked requests"),
	}
}

func (IPBan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("pattern").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("status"),
		index.Fields("expires_at"),
		index.Fields("last_hit_at"),
	}
}
