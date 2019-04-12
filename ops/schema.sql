CREATE TABLE public.manga (
	id serial NOT NULL,
	muid int NOT NULL,
	latest_release varchar NOT NULL,
	display_title VARCHAR NOT NULL,
	created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE INDEX manga_muid_idx (muid ASC),
	CONSTRAINT manga_pk PRIMARY KEY (id)
);

---

CREATE TABLE public.mangarelease (
	id serial NOT NULL,
	muid int NOT NULL,
	"release" varchar NOT NULL,
	translators varchar NOT NULL,
	created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT mangarelease_pk PRIMARY KEY (id),
	CONSTRAINT mangarelease_manga_fk FOREIGN KEY (muid) REFERENCES public.manga(muid) ON DELETE CASCADE ON UPDATE CASCADE,
	UNIQUE INDEX mangarelease_un (muid ASC, release ASC, translators ASC)
);

---
CREATE TABLE public.mangatitle (
	title varchar NOT NULL,
	muid int NOT NULL,
	CONSTRAINT mangatitles_pk PRIMARY KEY (title,muid),
	CONSTRAINT mangatitles_manga_fk FOREIGN KEY (muid) REFERENCES public.manga(muid) ON DELETE CASCADE ON UPDATE CASCADE
);

---
CREATE TABLE public.mangafeed (
	hash varchar NOT NULL,
	titles varchar[] NOT NULL,
	"type" varchar NOT NULL,
	created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT mangafeed_pk PRIMARY KEY (hash)
);
