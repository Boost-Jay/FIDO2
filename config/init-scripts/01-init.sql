-- 初始化WebAuthn所需的数据库表

-- 用户表
CREATE TABLE IF NOT EXISTS "user" (
    id VARCHAR(255) PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    challenge VARCHAR(255) NOT NULL,
    credential VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 初始测试用户
INSERT INTO user (user_name, display_name) VALUES
    ('testuser', 'Test User')
ON CONFLICT DO NOTHING;