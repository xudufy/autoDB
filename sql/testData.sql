
use autodb;


INSERT INTO users(email, pw, username)
VALUES ('km19@gmail.com','1234567890','km19');

INSERT INTO projects(pname, pw)
VALUES ('project1', 'apijf');

INSERT INTO project_developer(uid, pid, privilege)
VALUES (1, 1, 'owner'); /* assuming the auto_increment starts at 1 */

INSERT INTO tables(pid, name)
VALUES (1, 'table1'); /* assuming the auto_increment starts at 1 */

INSERT INTO apis(aid, tid, name, type, tmpl)
VALUES ('apiIDfromsha256', 1,'api1','public','1234567890');


drop table if exists nullTest;
create table nullTest (
    id BIGINT primary key ,
    intid INT default 14,
    time DATETIME default CURRENT_TIMESTAMP,
    nulltime DATETIME default null,
    nullable int default null,
    nullstring varchar(60) default null
);

insert into nullTest (id, time, nullstring) values (2147483649, '2020-03-28 08:00:00', '?? ?');
insert into nullTest (id, time) values (1, '2020-03-29 08:00:00');