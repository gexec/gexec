<script setup lang="ts">
import {
  columnFilteringFeature,
  createFilteredRowModel,
  FlexRender,
  tableFeatures,
  useTable,
  type ColumnDef,
  type ColumnFiltersState,
} from '@tanstack/vue-table'
import type { Repository, Credential } from '../../../../client'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { computed, h, ref, unref } from 'vue'
import { valueUpdater } from '@/lib/utils'
import { useRepositories } from '../../providers/RepositoriesProvider'
import DialogNewRepository from '../DialogNewRepository'

const { repositories, loadRepositories } = useRepositories()

const features = tableFeatures({
  columnFilteringFeature,
  filteredRowModel: createFilteredRowModel(),
})

const columns: ColumnDef<typeof features, Repository>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: ({ row }) => row.getValue('name'),
  },
  {
    accessorKey: 'slug',
    header: 'Slug',
    cell: ({ row }) => row.getValue('slug'),
  },
  {
    accessorKey: 'url',
    header: 'URL',
    cell: ({ row }) => row.getValue('url'),
  },
  {
    accessorKey: 'branch',
    header: 'Branch',
    cell: ({ row }) => row.getValue('branch'),
  },
  {
    accessorKey: 'credential',
    header: 'Credential',
    cell: ({ row }) => {
      const credential = row.getValue<Credential>('credential')

      if (!credential) {
        return h('span', {}, '—')
      }

      return h('span', {}, credential.name)
    },
  },
  {
    accessorKey: 'updated_at',
    header: () => h('div', { class: 'text-end' }, 'Last update'),
    cell: ({ row }) => {
      const formattedDate = new Intl.DateTimeFormat('en-US').format(
        new Date(row.getValue('updated_at'))
      )

      return h('div', { class: 'text-end' }, formattedDate)
    },
  },
]

const columnFilters = ref<ColumnFiltersState>([])

const table = useTable({
  features,
  // Using data directly without a getter was not reactive...
  // Might be due to the data coming from context
  // Not really worth the effort to investigate further since the getter works fine
  get data() {
    return unref(repositories)
  },
  columns,
  onColumnFiltersChange: (updaterOrValue) =>
    valueUpdater(updaterOrValue, columnFilters),
  state: {
    get columnFilters() {
      return columnFilters.value
    },
  },
})

const filter = computed({
  get() {
    return table.getColumn('name')?.getFilterValue() as string
  },
  set(value: string) {
    table.getColumn('name')?.setFilterValue(value)
  },
})

loadRepositories()
</script>

<template>
  <div class="w-full">
    <div class="flex gap-2 items-center justify-between py-4">
      <Input
        v-model="filter"
        class="max-w-sm"
        placeholder="Filter repositories by name"
      />
      <DialogNewRepository />
    </div>
    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow
            v-for="headerGroup in table.getHeaderGroups()"
            :key="headerGroup.id"
          >
            <TableHead v-for="header in headerGroup.headers" :key="header.id">
              <FlexRender v-if="!header.isPlaceholder" :header="header" />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <template v-if="table.getRowModel().rows?.length">
            <template v-for="row in table.getRowModel().rows" :key="row.id">
              <TableRow>
                <TableCell v-for="cell in row.getAllCells()" :key="cell.id">
                  <FlexRender :cell="cell" />
                </TableCell>
              </TableRow>
            </template>
          </template>

          <TableRow v-else>
            <TableCell :colspan="columns.length" class="h-24 text-center">
              No results.
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
