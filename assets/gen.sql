CREATE TABLE drug (
  yj_code         TEXT PRIMARY KEY,
  name            TEXT NOT NULL,
  name_normalized TEXT NOT NULL
);
INSERT INTO drug VALUES
 ('2189018F1043','エゼチミブ錠10mg「JG」','エゼチミブ錠10mgjg'),
 ('2189018F1094','エゼチミブ錠10mg「YD」','エゼチミブ錠10mgyd'),
 ('1149037F2093','セレコキシブ錠100mg「サワイ」','セレコキシブ錠100mgサワイ');
CREATE INDEX idx_norm ON drug(name_normalized);
