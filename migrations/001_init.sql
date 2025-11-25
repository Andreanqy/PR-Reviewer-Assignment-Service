CREATE TABLE teams (
    team_name VARCHAR(50) UNIQUE NOT NULL,
    members TEXT[]
);

CREATE TABLE users (
    user_id VARCHAR(50) UNIQUE NOT NULL,
    username VARCHAR(50),
    team_name VARCHAR(100),
    is_active boolean
);

CREATE TABLE pull_requests (
    pull_request_id VARCHAR(50),
    pull_request_name VARCHAR(100),
    author_id VARCHAR(50),
    "status" VARCHAR(20),
    assigned_reviewers TEXT[],
    createdAt TIMESTAMP,
    mergedAt TIMESTAMP
)