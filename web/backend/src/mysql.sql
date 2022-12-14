--IF db_id('casbin') IS NULL
CREATE DATABASE casbin;
GO
USE casbin;
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