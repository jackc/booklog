alter table books rename column media to format;

update books set format='text' where format='book';
update books set format='audio' where format='audiobook';

---- create above / drop below ----

update books set format='book' where format='text';
update books set format='audiobook' where format='audio';

alter table books rename column format to media;
