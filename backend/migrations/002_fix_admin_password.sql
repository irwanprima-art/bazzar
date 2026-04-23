-- Fix admin password hash (was placeholder)
UPDATE users SET password_hash = '$2a$10$mzcJN9DWrF1RvP8rrkbww.aMb5sLtYbWzdX431I19tNxrQarZySO6'
WHERE username = 'admin';
