select user_id::text, language_code, computed_score::text
from logs
where user_id='11111111-1111-4111-8111-111111111111' and language_code='zzz';
