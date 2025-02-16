<script setup lang="ts">
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useForm } from 'vee-validate'
import { computed, ref, unref, watch } from 'vue'
import { createProject } from '../../../../../client'
import { useProjectsStore } from '../../store/projects'
import { toast } from '@/components/ui/toast'

const { addProject } = useProjectsStore()

const formSchema = toTypedSchema(
  z.object({
    name: z.string().min(3).max(255),
    slug: z.string().min(3).max(255),
  })
)

const { handleSubmit, isSubmitting, isValidating, errors, values, setFieldValue, setFieldError } = useForm({
  validationSchema: formSchema,
})

const isOpen = ref(false)

const isSubmitDisabled = computed(() => {
  return (
    unref(isSubmitting) ||
    unref(isValidating) ||
    unref(errors).name ||
    unref(errors).slug ||
    !values.name ||
    !values.slug
  )
})

function closeModal() {
  isOpen.value = false
}

const onSubmit = handleSubmit(async (values) => {
  try {
    const { data, error } = await createProject({
      body: values as Required<typeof values>,
    })

    if (error?.status === 422) {
      for (const e of error.errors!) {
        setFieldError(e.field as 'name' | 'slug', e.message)
      }

      return
    }

    if (error) {
      console.error(error)
      toast({
        title: 'An error occurred',
        description:
          'We have encountered an unexpected issue while creating the project. Try again later.',
        variant: 'destructive',
      })

      return
    }

    addProject(data)
    closeModal()
    toast({
        title: 'Success',
        description:
          'Project successfully created.',
      })
  } catch (error) {
    console.error(error)
    toast({
        title: 'An error occurred',
        description:
          'We have encountered an unexpected issue while creating the project. Try again later.',
        variant: 'destructive',
      })
  }
})

function formatSlug(value: string): string {
  return value.replaceAll(' ', '-').toLowerCase()
}

watch(() => values.slug, (value) => {
  if (!value) {
    return
  }

  setFieldValue('slug', formatSlug(value))
})
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogTrigger as-child>
      <Button variant="outline">New</Button>
    </DialogTrigger>
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <DialogTitle>New project</DialogTitle>
        <DialogDescription class="sr-only">
          Enter project details. Click on Create when you're done.
        </DialogDescription>
      </DialogHeader>

      <form class="grid gap-4" @submit="onSubmit">
        <FormField v-slot="{ componentField }" name="name">
          <FormItem>
            <FormLabel>Name</FormLabel>
            <FormControl>
              <Input type="text" autocomplete="off" v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="slug">
          <FormItem>
            <FormLabel>Slug</FormLabel>
            <FormControl>
              <Input type="text" autocomplete="off" v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <DialogFooter>
          <Button variant="secondary" :disabled="isSubmitting" @click="closeModal">Cancel</Button>
          <Button type="submit" :disabled="isSubmitDisabled">Create</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
