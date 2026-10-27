select
  permission_relationships.object_type,
  permission_relationships.object_id,
  permission_relationships.relation
from permission_relationships
where permission_relationships.tenant = 'tadoku/prod'
  and permission_relationships.subject_type = 'user'
  and permission_relationships.subject_id = '11111111-1111-4111-8111-111111111111'
order by permission_relationships.object_type, permission_relationships.object_id, permission_relationships.relation
