select users.display_name
from users
where users.tenant = 'tadoku/prod'
  and users.id = '11111111-1111-4111-8111-111111111111'
order by users.id
