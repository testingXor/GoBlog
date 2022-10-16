alter table posts add visibility text not null default '';
update posts set status = 'published', visibility = 'public' where status = 'published';
update posts set status = 'published-deleted', visibility = 'public' where status = 'published-deleted';
update posts set status = 'draft', visibility = 'public' where status = 'draft';
update posts set status = 'draft-deleted', visibility = 'public' where status = 'draft-deleted';
update posts set status = 'scheduled', visibility = 'public' where status = 'scheduled';
update posts set status = 'scheduled-deleted', visibility = 'public' where status = 'scheduled-deleted';
update posts set status = 'published', visibility = 'private' where status = 'private';
update posts set status = 'published-deleted', visibility = 'private' where status = 'private-deleted';
update posts set status = 'published', visibility = 'unlisted' where status = 'unlisted';
update posts set status = 'published-deleted', visibility = 'unlisted' where status = 'unlisted-deleted';