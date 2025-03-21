import { defineCollection } from 'astro:content';
import { docsLoader } from '@astrojs/starlight/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

export const collections: any = {}
if(import.meta.env.DEV) {
	collections.docs = defineCollection({ loader: docsLoader(), schema: docsSchema() })
}
