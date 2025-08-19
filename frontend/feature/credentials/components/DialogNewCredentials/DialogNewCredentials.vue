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
import { ref, unref, watch } from 'vue'
import { createProjectCredential } from '../../../../client'
import { toast } from '@/components/ui/toast'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useProjectsStore } from '@/feature/projects/store/projects'
import { Textarea } from '@/components/ui/textarea'
import { useCredentials } from '../../providers/CredentialsProvider'

const projectsStore = useProjectsStore()
const { addCredentials } = useCredentials()

const formSchema = z.object({
  name: z.string().min(3).max(255),
  slug: z.string().min(3).max(255),
  kind: z.enum(['empty', 'shell', 'login']),
  shell: z
    .object({
      username: z.string().min(3).max(255).optional(),
      password: z.string().min(3).max(255).optional(),
      private_key: z.string().min(3).max(255).optional(),
    })
    .optional(),
  login: z
    .object({
      username: z.string().min(3).max(255).optional(),
      password: z.string().min(3).max(255).optional(),
    })
    .optional(),
})

const {
  handleSubmit,
  isSubmitting,
  isValidating,
  values,
  setFieldValue,
  setFieldError,
} = useForm({
  validationSchema: formSchema,
  initialValues: {
    kind: 'empty',
  },
})

const isOpen = ref(false)

function closeModal() {
  isOpen.value = false
}

const onSubmit = handleSubmit(async (values) => {
  try {
    const projectId = unref(projectsStore.selectedProject)?.id

    if (!projectId) {
      toast({
        title: 'An error occurred',
        description: 'No project selected.',
        variant: 'destructive',
      })

      return
    }

    let hasError = false

    if (values.kind === 'shell') {
      if (!values.shell?.private_key) {
        setFieldError('shell.private_key', 'Private key is required.')
        hasError = true
      }

      if (!values.shell?.username) {
        setFieldError('shell.username', 'Username is required.')
        hasError = true
      }

      if (!values.shell?.password) {
        setFieldError('shell.password', 'Password is required.')
        hasError = true
      }
    }

    if (values.kind === 'login') {
      if (!values.login?.username) {
        setFieldError('login.username', 'Username is required.')
        hasError = true
      }

      if (!values.login?.password) {
        setFieldError('login.password', 'Password is required.')
        hasError = true
      }
    }

    if (hasError) {
      return
    }

    const { data, error } = await createProjectCredential({
      path: {
        project_id: projectId,
      },
      body: {
        ...values,
        shell: values.kind === 'shell' ? values.shell : undefined,
        login: values.kind === 'login' ? values.login : undefined,
      },
    })

    if (error?.status === 422) {
      for (const e of error.errors!) {
        setFieldError(
          // @ts-expect-error would be cumbersome to type the field to the keys
          e.field,
          e.message
        )
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

    addCredentials(data)
    closeModal()
    toast({
      title: 'Success',
      description: 'Credentials successfully created.',
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

watch(
  () => values.slug,
  (value) => {
    if (!value) {
      return
    }

    setFieldValue('slug', formatSlug(value))
  }
)

watch(isOpen, (value) => {
  if (!value) {
    return
  }

  setFieldValue('kind', 'empty')
})
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogTrigger as-child>
      <Button variant="outline">New</Button>
    </DialogTrigger>
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <DialogTitle>New credentials</DialogTitle>
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

        <FormField v-slot="{ componentField }" name="kind">
          <FormItem>
            <FormLabel>Kind</FormLabel>
            <Select v-bind="componentField">
              <FormControl>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="empty">Empty</SelectItem>
                  <SelectItem value="shell">Shell</SelectItem>
                  <SelectItem value="login">Login</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <FormMessage />
          </FormItem>
        </FormField>

        <template v-if="values.kind === 'login'">
          <FormField v-slot="{ componentField }" name="login.username">
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input type="text" autocomplete="off" v-bind="componentField" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="login.password">
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <Input
                  type="password"
                  autocomplete="off"
                  v-bind="componentField"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </template>

        <template v-if="values.kind === 'shell'">
          <FormField v-slot="{ componentField }" name="shell.username">
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input type="text" autocomplete="off" v-bind="componentField" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="shell.password">
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <Input
                  type="password"
                  autocomplete="off"
                  v-bind="componentField"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="shell.private_key">
            <FormItem>
              <FormLabel>Private key</FormLabel>
              <FormControl>
                <Textarea v-bind="componentField" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </template>

        <DialogFooter>
          <Button
            variant="secondary"
            :disabled="isSubmitting || isValidating"
            @click="closeModal"
            >Cancel</Button
          >
          <Button type="submit" :disabled="isSubmitting || isValidating"
            >Create</Button
          >
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
