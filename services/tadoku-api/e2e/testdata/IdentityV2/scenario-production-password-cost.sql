update accounts
set password_hash = '$2b$12$TP8U/FQ8vCC/fzTCu5Fgve4WAHMgvG1jZ7V7bKWjDJRA.KeeC1xqK'
where email_normalized = 'reader@example.test';
