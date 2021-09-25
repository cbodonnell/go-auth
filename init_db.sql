-- Initialization Script


-- public.users definition
DO $$
BEGIN
	CREATE TABLE IF NOT EXISTS public.users (
		id serial NOT NULL,
		username varchar NOT NULL,
		"password" varchar NOT NULL,
		created timestamptz NOT NULL,
		uuid text NOT NULL,
		CONSTRAINT users_pkey PRIMARY KEY (id)
	);


	-- public."groups" definition

	CREATE TABLE IF NOT EXISTS public."groups" (
		id serial NOT NULL,
		"name" varchar NOT NULL,
		CONSTRAINT groups_pkey PRIMARY KEY (id)
	);

	IF NOT EXISTS (SELECT * FROM public.groups) THEN
		INSERT INTO public."groups" ("name") VALUES ('public'), ('admin');
	END IF;


	-- public.user_refresh definition

	CREATE TABLE IF NOT EXISTS public.user_refresh (
		id serial NOT NULL,
		user_id int4 NOT NULL,
		"jti" text NOT NULL,
		expires timestamptz NOT NULL,
		CONSTRAINT user_refresh_pkey PRIMARY KEY (id)
	);

	-- public.user_refresh foreign keys
	ALTER TABLE public.user_refresh DROP CONSTRAINT IF EXISTS fki_user_refresh_user_id;
	ALTER TABLE public.user_refresh ADD CONSTRAINT fki_user_refresh_user_id FOREIGN KEY (user_id) REFERENCES public.users(id);


	-- public."user_groups" definition

	CREATE TABLE IF NOT EXISTS public.user_groups (
		id serial NOT NULL,
		user_id int4 NOT NULL,
		group_id int4 NOT NULL,
		CONSTRAINT user_groups_pkey PRIMARY KEY (id)
	);
	CREATE INDEX IF NOT EXISTS fki_user_groups_group_id ON public.user_groups USING btree (group_id);
	CREATE INDEX IF NOT EXISTS fki_user_groups_user_id ON public.user_groups USING btree (user_id);

	-- public.user_groups foreign keys

	ALTER TABLE public.user_groups DROP CONSTRAINT IF EXISTS user_groups_group_id;
	ALTER TABLE public.user_groups ADD CONSTRAINT user_groups_group_id FOREIGN KEY (group_id) REFERENCES public."groups"(id);
	ALTER TABLE public.user_groups DROP CONSTRAINT IF EXISTS user_groups_user_id;
	ALTER TABLE public.user_groups ADD CONSTRAINT user_groups_user_id FOREIGN KEY (user_id) REFERENCES public.users(id);
END
$$


