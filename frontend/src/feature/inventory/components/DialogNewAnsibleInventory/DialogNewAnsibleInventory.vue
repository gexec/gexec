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
import { Button } from '@/components/ui/button'
import { useForm } from 'vee-validate'
import { ref, shallowRef, unref, watch } from 'vue'
import { createProjectRepository } from '../../../../../client'
import { toast } from '@/components/ui/toast'
import { useProjectsStore } from '@/feature/projects/store/projects'
import { AutoForm } from '@/components/ui/auto-form'
import { useCredentials } from '@/feature/credentials/providers/CredentialsProvider'
import { Skeleton } from '@/components/ui/skeleton'
import { useInventories } from '../../providers/InventoriesProvider'
import { useRepositories } from '@/feature/repositories/providers/RepositoriesProvider'
import { formatSlug } from '@/lib/utils'

const projectsStore = useProjectsStore()
const { addInventory } = useInventories()
const { credentials, loadCredentials } = useCredentials()
const { loadRepositories, repositories } = useRepositories()

const defaultSchema = Object.freeze({
  name: z.string().min(3).max(255),
  slug: z.string().min(3).max(255),
})

const formSchema = shallowRef()

const { isSubmitting, isValidating, values, setFieldValue, setFieldError } =
  useForm({
    validationSchema: formSchema,
  })

const isOpen = ref(false)

function closeModal() {
  isOpen.value = false
}

async function onSubmit(values: Record<string, unknown>) {
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

    const credential = credentials.value.find(
      (c) => c.name === values.credential_id
    )

    if (!credential) {
      setFieldError('credential_id', 'Credential not found')
      return
    }

    const { data, error } = await createProjectRepository({
      path: {
        project_id: projectId,
      },
      body: { ...values, credential_id: credential.id },
    })

    if (error?.status === 422) {
      for (const e of error.errors!) {
        setFieldError(e.field!, e.message)
      }

      return
    }

    if (error) {
      console.error(error)
      toast({
        title: 'An error occurred',
        description:
          'We have encountered an unexpected issue while creating the inventory. Try again later.',
        variant: 'destructive',
      })

      return
    }

    addInventory({ ...data, credential })
    closeModal()
    toast({
      title: 'Success',
      description: 'Inventory successfully created.',
    })
  } catch (error) {
    console.error(error)
    toast({
      title: 'An error occurred',
      description:
        'We have encountered an unexpected issue while creating the inventory. Try again later.',
      variant: 'destructive',
    })
  }
}

async function loadDependencies() {
  await Promise.allSettled([loadCredentials(), loadRepositories()])

  const credentialNames = credentials.value.map((c) => c.name!)
  const repositoryNames = repositories.value.map((r) => r.name!)

  formSchema.value = z.object({
    ...defaultSchema,
    credential_id: z.enum([
      credentialNames.at(0)!,
      ...credentialNames.slice(1),
    ]),
    repository_id: z.enum([
      repositoryNames.at(0)!,
      ...repositoryNames.slice(1),
    ]),
  })
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

  loadDependencies()
})
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogTrigger as-child>
      <Button variant="outline">New</Button>
    </DialogTrigger>
    <DialogContent class="max-w-lg">
      <DialogHeader>
        <DialogTitle>New Ansible inventory</DialogTitle>
        <DialogDescription class="sr-only">
          Enter project details. Click on Create when you're done.
        </DialogDescription>
      </DialogHeader>

      <template v-if="isLoadingCredentials || isLoadingRepositories">
        <div>
          <Skeleton class="h-3.5 w-14 my-[1.5px]" />
          <Skeleton class="h-10 w-full mt-2" />
        </div>

        <div>
          <Skeleton class="h-3.5 w-14 my-[1.5px]" />
          <Skeleton class="h-10 w-full mt-2" />
        </div>

        <div>
          <Skeleton class="h-3.5 w-14 my-[1.5px]" />
          <Skeleton class="h-10 w-full mt-2" />
        </div>

        <div>
          <Skeleton class="h-3.5 w-14 my-[1.5px]" />
          <Skeleton class="h-10 w-full mt-2" />
        </div>

        <div>
          <Skeleton class="h-3.5 w-14 my-[1.5px]" />
          <Skeleton class="h-10 w-full mt-2" />
        </div>
      </template>

      <AutoForm
        v-else
        :schema="formSchema"
        class="grid gap-4"
        @submit="onSubmit"
      >
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
      </AutoForm>
    </DialogContent>
  </Dialog>
</template>
