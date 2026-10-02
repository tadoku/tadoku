select tenant, count(*) as count
from logs
group by tenant
order by tenant;
