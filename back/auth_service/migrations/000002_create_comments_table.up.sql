CREATE TABLE comments(
                         id BIGSERIAL PRIMARY KEY,
                         user_id BIGINT NOT NULL,
                         post_id BIGINT NOT NULL,
                         content TEXT NOT NULL,
                         created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                         FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
)