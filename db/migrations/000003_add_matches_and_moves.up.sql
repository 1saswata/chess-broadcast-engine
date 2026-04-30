CREATE TABLE IF NOT EXISTS matches (
    id VARCHAR(60) PRIMARY KEY,
    white_player_id UUID REFERENCES users(id),
    black_player_id UUID REFERENCES users(id),
    status VARCHAR(30),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS moves (
    id UUID PRIMARY KEY,
    match_id VARCHAR(60) REFERENCES matches(id),
    sequence_number INTEGER,
    move_payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP 
)