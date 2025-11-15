package models

type Chat struct {
	TgID      int64 `pg:",pk,notnull"`
	CreatedAt int64 `pg:"default:extract(epoch from now())"`
	UpdatedAt int64 `pg:"default:extract(epoch from now())"`

	ChatOwnerID int64 `pg:",notnull"`
	ChatOwner   *User `pg:"rel:has-one,fk:chat_owner_id"`

	AssociatedThreads []*Thread `pg:"rel:has-many,join_fk:chat_id"`
}
