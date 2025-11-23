CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE teams (
    team_name VARCHAR PRIMARY KEY
);

CREATE TABLE users (
    user_id VARCHAR PRIMARY KEY,
    username VARCHAR NOT NULL,
    team_name VARCHAR NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE pull_requests (
    pull_request_id VARCHAR PRIMARY KEY,
    pull_request_name VARCHAR NOT NULL,
    author_id VARCHAR NOT NULL REFERENCES users(user_id),
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE pull_request_shorts (
    pull_request_id VARCHAR NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    author_id VARCHAR NOT NULL REFERENCES users(user_id),
    PRIMARY KEY (pull_request_id, author_id)
);
