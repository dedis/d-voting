-- Create database 'casbin' if not exists
-- https://stackoverflow.com/a/36591842
SELECT 'CREATE DATABASE casbin'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'casbin')\gexec

-- Connect to the casbin DB
\c casbin

CREATE TABLE IF NOT EXISTS casbin (
    id SERIAL PRIMARY KEY,
    ptype varchar(256) default ''::character varying not null,
    rule  text[]
  );

alter table casbin
    owner to dvoting;

create index idx_casbin_id
    on casbin(id);

create index idx_casbin_ptype
    on casbin(ptype);

create index idx_casbin_rule
    on casbin(rule);

INSERT INTO casbin(id,ptype, rule) VALUES (1,'p', ARRAY['330383', 'roles', 'list']);
INSERT INTO casbin(id,ptype, rule) VALUES (2,'p', ARRAY['330383', 'roles', 'remove']);
INSERT INTO casbin(id,ptype, rule) VALUES (3,'p', ARRAY['330383', 'roles', 'add']);
INSERT INTO casbin(id,ptype, rule) VALUES (4,'p', ARRAY['330383', 'proxies', 'post']);
INSERT INTO casbin(id,ptype, rule) VALUES (5,'p', ARRAY['330383', 'proxies', 'put']);
INSERT INTO casbin(id,ptype, rule) VALUES (6,'p', ARRAY['330383', 'proxies', 'delete']);
INSERT INTO casbin(id,ptype, rule) VALUES (7,'p', ARRAY['330383', 'election', 'create']);

INSERT INTO casbin(id,ptype, rule) VALUES (8,'p', ARRAY['330382', 'roles', 'list']);
INSERT INTO casbin(id,ptype, rule) VALUES (9,'p', ARRAY['330382', 'roles', 'remove']);
INSERT INTO casbin(id,ptype, rule) VALUES (10,'p', ARRAY['330382', 'roles', 'add']);
INSERT INTO casbin(id,ptype, rule) VALUES (11,'p', ARRAY['330382', 'proxies', 'post']);
INSERT INTO casbin(id,ptype, rule) VALUES (12,'p', ARRAY['330382', 'proxies', 'put']);
INSERT INTO casbin(id,ptype, rule) VALUES (13,'p', ARRAY['330382', 'proxies', 'delete']);
INSERT INTO casbin(id,ptype, rule) VALUES (14,'p', ARRAY['330382', 'election', 'create']);

INSERT INTO casbin(id,ptype, rule) VALUES (15,'p', ARRAY['228271', 'roles', 'list']);
INSERT INTO casbin(id,ptype, rule) VALUES (16,'p', ARRAY['228271', 'roles', 'remove']);
INSERT INTO casbin(id,ptype, rule) VALUES (17,'p', ARRAY['228271', 'roles', 'add']);
INSERT INTO casbin(id,ptype, rule) VALUES (18,'p', ARRAY['228271', 'proxies', 'post']);
INSERT INTO casbin(id,ptype, rule) VALUES (19,'p', ARRAY['228271', 'proxies', 'put']);
INSERT INTO casbin(id,ptype, rule) VALUES (20,'p', ARRAY['228271', 'proxies', 'delete']);
INSERT INTO casbin(id,ptype, rule) VALUES (21,'p', ARRAY['228271', 'election', 'create']);

INSERT INTO casbin(id,ptype, rule) VALUES (22,'p', ARRAY['330361', 'roles', 'list']);
INSERT INTO casbin(id,ptype, rule) VALUES (23,'p', ARRAY['330361', 'roles', 'remove']);
INSERT INTO casbin(id,ptype, rule) VALUES (24,'p', ARRAY['330361', 'roles', 'add']);
INSERT INTO casbin(id,ptype, rule) VALUES (25,'p', ARRAY['330361', 'proxies', 'post']);
INSERT INTO casbin(id,ptype, rule) VALUES (26,'p', ARRAY['330361', 'proxies', 'put']);
INSERT INTO casbin(id,ptype, rule) VALUES (27,'p', ARRAY['330361', 'proxies', 'delete']);
INSERT INTO casbin(id,ptype, rule) VALUES (28,'p', ARRAY['330361', 'election', 'create']);

INSERT INTO casbin(id,ptype, rule) VALUES (29,'p', ARRAY['175129', 'roles', 'list']);
INSERT INTO casbin(id,ptype, rule) VALUES (30,'p', ARRAY['175129', 'roles', 'remove']);
INSERT INTO casbin(id,ptype, rule) VALUES (31,'p', ARRAY['175129', 'roles', 'add']);
INSERT INTO casbin(id,ptype, rule) VALUES (32,'p', ARRAY['175129', 'proxies', 'post']);
INSERT INTO casbin(id,ptype, rule) VALUES (33,'p', ARRAY['175129', 'proxies', 'put']);
INSERT INTO casbin(id,ptype, rule) VALUES (34,'p', ARRAY['175129', 'proxies', 'delete']);
INSERT INTO casbin(id,ptype, rule) VALUES (35,'p', ARRAY['175129', 'election', 'create']);

INSERT INTO casbin(id,ptype, rule) VALUES (36,'p', ARRAY['324610', 'roles', 'list']);
INSERT INTO casbin(id,ptype, rule) VALUES (37,'p', ARRAY['324610', 'roles', 'remove']);
INSERT INTO casbin(id,ptype, rule) VALUES (38,'p', ARRAY['324610', 'roles', 'add']);
INSERT INTO casbin(id,ptype, rule) VALUES (39,'p', ARRAY['324610', 'proxies', 'post']);
INSERT INTO casbin(id,ptype, rule) VALUES (40,'p', ARRAY['324610', 'proxies', 'put']);
INSERT INTO casbin(id,ptype, rule) VALUES (41,'p', ARRAY['324610', 'proxies', 'delete']);
INSERT INTO casbin(id,ptype, rule) VALUES (42,'p', ARRAY['324610', 'election', 'create']);

INSERT INTO casbin(id,ptype, rule) VALUES (43,'p', ARRAY['315822', 'roles', 'list']);
INSERT INTO casbin(id,ptype, rule) VALUES (44,'p', ARRAY['315822', 'roles', 'remove']);
INSERT INTO casbin(id,ptype, rule) VALUES (45,'p', ARRAY['315822', 'roles', 'add']);
INSERT INTO casbin(id,ptype, rule) VALUES (46,'p', ARRAY['315822', 'proxies', 'post']);
INSERT INTO casbin(id,ptype, rule) VALUES (47,'p', ARRAY['315822', 'proxies', 'put']);
INSERT INTO casbin(id,ptype, rule) VALUES (48,'p', ARRAY['315822', 'proxies', 'delete']);
INSERT INTO casbin(id,ptype, rule) VALUES (49,'p', ARRAY['315822', 'election', 'create']);

INSERT INTO casbin(id,ptype, rule) VALUES (50,'p', ARRAY['321016', 'roles', 'list']);
INSERT INTO casbin(id,ptype, rule) VALUES (51,'p', ARRAY['321016', 'roles', 'remove']);
INSERT INTO casbin(id,ptype, rule) VALUES (52,'p', ARRAY['321016', 'roles', 'add']);
INSERT INTO casbin(id,ptype, rule) VALUES (53,'p', ARRAY['321016', 'proxies', 'post']);
INSERT INTO casbin(id,ptype, rule) VALUES (54,'p', ARRAY['321016', 'proxies', 'put']);
INSERT INTO casbin(id,ptype, rule) VALUES (55,'p', ARRAY['321016', 'proxies', 'delete']);   
INSERT INTO casbin(id,ptype, rule) VALUES (56,'p', ARRAY['321016', 'election', 'create']);
