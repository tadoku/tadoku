-- name: RelationshipSubjects :many
select subject_type, subject_id, subject_relation
from permission_relationships
where object_type = @object_type
  and object_id = @object_id
  and relation = @relation
  and (subject_type <> 'user' or subject_id = @user_id);

-- name: InsertRelationship :execrows
insert into permission_relationships (
  object_type, object_id, relation, subject_type, subject_id, subject_relation
)
values (@object_type, @object_id, @relation, @subject_type, @subject_id, @subject_relation)
on conflict do nothing;

-- name: DeleteRelationship :execrows
delete from permission_relationships
where object_type = @object_type
  and object_id = @object_id
  and relation = @relation
  and subject_type = @subject_type
  and subject_id = @subject_id
  and subject_relation = @subject_relation;

-- name: UserRelations :many
select relation, subject_id
from permission_relationships
where object_type = 'app'
  and object_id = 'root'
  and relation = any(@relations::text[])
  and subject_type = 'user'
  and subject_id = any(@user_ids::text[]);
