<script setup lang="ts">
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  columnFilteringFeature,
  createFilteredRowModel,
  FlexRender,
  tableFeatures,
  useTable,
  type ColumnDef,
  type ColumnFiltersState,
} from '@tanstack/vue-table'
import { h, ref } from 'vue'
import { useProjectsStore } from '../../store/projects'
import { storeToRefs } from 'pinia'
import type { Project } from '../../../../client'
import DialogNewProject from '../DialogNewProject'
import { Input } from '@/components/ui/input'
import { valueUpdater } from '@/lib/utils'

const projectsStore = useProjectsStore()
const { projects } = storeToRefs(projectsStore)

const features = tableFeatures({
  columnFilteringFeature,
  filteredRowModel: createFilteredRowModel(),
})

const columns: ColumnDef<typeof features, Project>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: ({ row }) => h('div', { class: 'font-medium' }, row.getValue('name')),
  },
  {
    accessorKey: 'slug',
    header: 'Slug',
    cell: ({ row }) => row.getValue('slug'),
  },
  {
    accessorKey: 'updated_at',
    header: () => h('div', { class: 'text-end' }, 'Last update'),
    cell: ({ row }) => {
      return h(
        'div',
        { class: 'text-end' },
        new Intl.DateTimeFormat('en-US').format(
          new Date(row.getValue('updated_at'))
        )
      )
    },
  },
]

const columnFilters = ref<ColumnFiltersState>([])

const table = useTable({
  features,
  data: projects,
  columns,
  onColumnFiltersChange: (updaterOrValue) =>
    valueUpdater(updaterOrValue, columnFilters),
  state: {
    get columnFilters() {
      return columnFilters.value
    },
  },
})
</script>

<template>
  <div class="w-full">
    <div class="flex gap-2 items-center justify-between py-4">
      <Input
        class="max-w-sm"
        placeholder="Filter projects by name"
        :model-value="table.getColumn('name')?.getFilterValue() as string"
        @update:model-value="table.getColumn('name')?.setFilterValue($event)"
      />
      <DialogNewProject />
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
