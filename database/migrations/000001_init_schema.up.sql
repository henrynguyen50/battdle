-- Create divisions table
CREATE TABLE divisions (
    id SERIAL PRIMARY KEY,
    league VARCHAR(10) NOT NULL,
    name VARCHAR(50) NOT NULL UNIQUE
);

-- Create teams table
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    mlb_team_id INT NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    division_id INT REFERENCES divisions(id)
);

-- Create players table
CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    mlb_id INT NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    birth_date DATE,
    birth_year INT NOT NULL,
    birth_city VARCHAR(100),
    birth_country VARCHAR(100),
    position VARCHAR(20) NOT NULL DEFAULT 'P',
    height VARCHAR(20),
    weight INT,
    mlb_debut_year INT NOT NULL,
    mlb_last_year INT NOT NULL,
    team_id INT REFERENCES teams(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create pitch_profiles table
CREATE TABLE pitch_profiles (
    id SERIAL PRIMARY KEY,
    player_id INT REFERENCES players(id) ON DELETE CASCADE,
    pitch_type VARCHAR(20),
    velocity FLOAT NOT NULL,
    spin_rate FLOAT NOT NULL,
    release_pos_x FLOAT NOT NULL,
    release_pos_z FLOAT NOT NULL,
    release_extension FLOAT NOT NULL,
    break_x FLOAT NOT NULL,
    break_z FLOAT NOT NULL,
    arm_angle FLOAT NOT NULL,
    plate_x FLOAT NOT NULL,
    plate_z FLOAT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create daily_puzzles table
CREATE TABLE daily_puzzles (
    id SERIAL PRIMARY KEY,
    puzzle_date DATE NOT NULL UNIQUE,
    target_player_id INT NOT NULL REFERENCES players(id),
    target_pitch_profile_id INT NOT NULL REFERENCES pitch_profiles(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create guesses table
CREATE TABLE guesses (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    puzzle_id INT NOT NULL REFERENCES daily_puzzles(id),
    guessed_player_id INT NOT NULL REFERENCES players(id),
    balls INT NOT NULL,
    strikes INT NOT NULL,
    result VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(session_id, puzzle_id, guessed_player_id)
);

-- Create animations table
CREATE TABLE animations (
    id SERIAL PRIMARY KEY,
    pitch_profile_id INT NOT NULL UNIQUE REFERENCES pitch_profiles(id) ON DELETE CASCADE,
    animation_data JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
