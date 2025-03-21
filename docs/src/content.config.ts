import { defineCollection } from 'astro:content';
import { docsLoader } from '@astrojs/starlight/loaders';
import { docsSchema } from '@astrojs/starlight/schema';
import {isDev} from "./dev.ts";

export const collections: any = {}
if(isDev()) {
	collections.docs = defineCollection({ loader: docsLoader(), schema: docsSchema() })
}
