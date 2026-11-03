update accounts
set password_hash = '$2b$12$TP8U/FQ8vCC/fzTCu5Fgve4WAHMgvG1jZ7V7bKWjDJRA.KeeC1xqK'
where tenant in (
  'tadoku/prod',
  'tadoku/identity-a-0123abcd',
  'tadoku/identity-b-0123abcd'
)
  and email_normalized = 'reader@example.test';
