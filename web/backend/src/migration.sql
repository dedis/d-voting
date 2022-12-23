CREATE DATABASE casbin
/*
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
INSERT INTO casbin(id,ptype, rule) VALUES (56,'p', ARRAY['321016', 'election', 'create']);*/


--IF db_id('casbin') IS NULL
--CREATE DATABASE casbin;
--GO
--USE casbin;
/*create table casbin
(
    p_type varchar(256) default ''::character varying not null,
    v0     varchar(256) default ''::character varying not null,
    v1     varchar(256) default ''::character varying not null,
    v2     varchar(256) default ''::character varying not null,
    v3     varchar(256) default ''::character varying not null,
    v4     varchar(256) default ''::character varying not null,
    v5     varchar(256) default ''::character varying not null
);

alter table casbin
    owner to dvoting;

create index idx_casbin_p_type
    on casbin(p_type);

create index idx_casbin_v0
    on casbin(v0);

create index idx_casbin_v1
    on casbin(v1);

create index idx_casbin_v2
    on casbin(v2);

create index idx_casbin_v3
    on casbin(v3);

create index idx_casbin_v4
    on casbin(v4);

create index idx_casbin_v5
    on casbin(v5);

INSERT INTO casbin(p_type, v0, v1, v2, v3, v4, v5) VALUES ('p', '330383', 'roles', 'list', '', '', '');
INSERT INTO casbin(p_type, v0, v1, v2, v3, v4, v5) VALUES ('p', '330383', 'roles', 'remove', '', '', '');
INSERT INTO casbin(p_type, v0, v1, v2, v3, v4, v5) VALUES ('p', '330383', 'roles', 'add', '', '', '');
INSERT INTO casbin(p_type, v0, v1, v2, v3, v4, v5) VALUES ('p', '330383', 'proxies', 'post', '', '', '');
INSERT INTO casbin(p_type, v0, v1, v2, v3, v4, v5) VALUES ('p', '330383', 'proxies', 'put', '', '', '');
INSERT INTO casbin(p_type, v0, v1, v2, v3, v4, v5) VALUES ('p', '330383', 'proxies', 'delete', '', '', '');
INSERT INTO casbin(p_type, v0, v1, v2, v3, v4, v5) VALUES ('p', '330383', 'election', 'create', '', '', '');*/