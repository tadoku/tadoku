-- name: RelationshipSubjects :many
select subject_type, subject_id, subject_relation
from permission_relationships
where object_type = @object_type
  and object_id = @object_id
  and relation = @relation
  and (subject_type <> 'user' or subject_id = @user_id);

-- name: InsertRelationships :execrows
insert into permission_relationships (
  object_type, object_id, relation, subject_type, subject_id, subject_relation
)
select
  unnest(@object_types::text[]),
  unnest(@object_ids::text[]),
  unnest(@relations::text[]),
  unnest(@subject_types::text[]),
  unnest(@subject_ids::text[]),
  unnest(@subject_relations::text[])
on conflict do nothing;

-- name: DeleteRelationships :execrows
delete from permission_relationships
where (object_type, object_id, relation, subject_type, subject_id, subject_relation) in (
  select
    unnest(@object_types::text[]),
    unnest(@object_ids::text[]),
    unnest(@relations::text[]),
    unnest(@subject_types::text[]),
    unnest(@subject_ids::text[]),
    unnest(@subject_relations::text[])
);

-- name: DeleteObjectRelationships :execrows
delete from permission_relationships
where (object_type = @object_type and object_id = @object_id)
  or (subject_type = @object_type and subject_id = @object_id);

-- name: DeleteUserRelationships :execrows
delete from permission_relationships
where subject_type = 'user'
  and subject_id = @user_id;

-- name: ReadRelationships :many
select object_type, object_id, relation, subject_type, subject_id, subject_relation
from permission_relationships
where object_type = @object_type::text
  and object_id = @object_id::text
  and @user_id::text = ''
  and (@relation::text = '' or relation = @relation::text)
union all
select object_type, object_id, relation, subject_type, subject_id, subject_relation
from permission_relationships
where subject_type = 'user'
  and subject_id = @user_id::text
  and @object_id::text = ''
  and (@object_type::text = '' or object_type = @object_type::text)
  and (@relation::text = '' or relation = @relation::text)
order by object_type, object_id, relation, subject_type, subject_id, subject_relation;

-- name: StoredRelationships :many
select object_type, object_id, relation, subject_type, subject_id, subject_relation
from permission_relationships
where (object_type, object_id, relation, subject_type, subject_id, subject_relation) in (
  select
    unnest(@object_types::text[]),
    unnest(@object_ids::text[]),
    unnest(@relations::text[]),
    unnest(@subject_types::text[]),
    unnest(@subject_ids::text[]),
    unnest(@subject_relations::text[])
);
