-- Seed divisions
INSERT INTO divisions (league, name) VALUES
('AL', 'AL East'),
('AL', 'AL Central'),
('AL', 'AL West'),
('NL', 'NL East'),
('NL', 'NL Central'),
('NL', 'NL West')
ON CONFLICT (name) DO NOTHING;

-- Seed teams
-- AL East
INSERT INTO teams (mlb_team_id, name, division_id) VALUES
(110, 'Baltimore Orioles', (SELECT id FROM divisions WHERE name = 'AL East')),
(111, 'Boston Red Sox', (SELECT id FROM divisions WHERE name = 'AL East')),
(147, 'New York Yankees', (SELECT id FROM divisions WHERE name = 'AL East')),
(139, 'Tampa Bay Rays', (SELECT id FROM divisions WHERE name = 'AL East')),
(141, 'Toronto Blue Jays', (SELECT id FROM divisions WHERE name = 'AL East')),
-- AL Central
(145, 'Chicago White Sox', (SELECT id FROM divisions WHERE name = 'AL Central')),
(114, 'Cleveland Guardians', (SELECT id FROM divisions WHERE name = 'AL Central')),
(116, 'Detroit Tigers', (SELECT id FROM divisions WHERE name = 'AL Central')),
(118, 'Kansas City Royals', (SELECT id FROM divisions WHERE name = 'AL Central')),
(142, 'Minnesota Twins', (SELECT id FROM divisions WHERE name = 'AL Central')),
-- AL West
(117, 'Houston Astros', (SELECT id FROM divisions WHERE name = 'AL West')),
(108, 'Los Angeles Angels', (SELECT id FROM divisions WHERE name = 'AL West')),
(133, 'Oakland Athletics', (SELECT id FROM divisions WHERE name = 'AL West')),
(136, 'Seattle Mariners', (SELECT id FROM divisions WHERE name = 'AL West')),
(140, 'Texas Rangers', (SELECT id FROM divisions WHERE name = 'AL West')),
-- NL East
(144, 'Atlanta Braves', (SELECT id FROM divisions WHERE name = 'NL East')),
(146, 'Miami Marlins', (SELECT id FROM divisions WHERE name = 'NL East')),
(121, 'New York Mets', (SELECT id FROM divisions WHERE name = 'NL East')),
(143, 'Philadelphia Phillies', (SELECT id FROM divisions WHERE name = 'NL East')),
(120, 'Washington Nationals', (SELECT id FROM divisions WHERE name = 'NL East')),
-- NL Central
(112, 'Chicago Cubs', (SELECT id FROM divisions WHERE name = 'NL Central')),
(113, 'Cincinnati Reds', (SELECT id FROM divisions WHERE name = 'NL Central')),
(158, 'Milwaukee Brewers', (SELECT id FROM divisions WHERE name = 'NL Central')),
(134, 'Pittsburgh Pirates', (SELECT id FROM divisions WHERE name = 'NL Central')),
(138, 'St. Louis Cardinals', (SELECT id FROM divisions WHERE name = 'NL Central')),
-- NL West
(109, 'Arizona Diamondbacks', (SELECT id FROM divisions WHERE name = 'NL West')),
(115, 'Colorado Rockies', (SELECT id FROM divisions WHERE name = 'NL West')),
(119, 'Los Angeles Dodgers', (SELECT id FROM divisions WHERE name = 'NL West')),
(135, 'San Diego Padres', (SELECT id FROM divisions WHERE name = 'NL West')),
(137, 'San Francisco Giants', (SELECT id FROM divisions WHERE name = 'NL West'))
ON CONFLICT (mlb_team_id) DO NOTHING;
