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
import { useUsers } from '../../providers/UsersProvider'
import type { User } from '../../../../client'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import DialogNewUser from '../DialogNewUser'
import { computed, h, ref, unref } from 'vue'
import { valueUpdater } from '@/lib/utils'

const { users, loadUsers } = useUsers()

const features = tableFeatures({
  columnFilteringFeature,
  filteredRowModel: createFilteredRowModel(),
})

const columns: ColumnDef<typeof features, User>[] = [
  {
    accessorKey: 'username',
    header: 'Username',
    cell: ({ row }) => row.getValue('username'),
  },
  {
    accessorKey: 'email',
    header: 'Email',
    cell: ({ row }) => row.getValue('email'),
  },
  {
    accessorKey: 'fullname',
    header: 'Fullnane',
    cell: ({ row }) => row.getValue('fullname'),
  },
  {
    accessorKey: 'active',
    header: 'Active',
    cell: ({ row }) => (row.getValue('active') ? 'Yes' : 'No'),
  },
  {
    accessorKey: 'admin',
    header: 'Admin',
    cell: ({ row }) => (row.getValue('admin') ? 'Yes' : 'No'),
  },
  {
    accessorKey: 'created_at',
    header: () => h('div', { class: 'text-end' }, 'Created'),
    cell: ({ row }) => {
      const formattedDate = new Intl.DateTimeFormat('en-US').format(
        new Date(row.getValue('created_at'))
      )

      return h('div', { class: 'text-end' }, formattedDate)
    },
  },
  {
    accessorKey: 'updated_at',
    header: () => h('div', { class: 'text-end' }, 'Updated'),
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
  get data() {
    return unref(users)
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
    return table.getColumn('username')?.getFilterValue() as string
  },
  set(value: string) {
    table.getColumn('username')?.setFilterValue(value)
  },
})

loadUsers()
</script>

<template>
  <div class="w-full">
    <div class="flex gap-2 items-center justify-between py-4">
      <Input
        v-model="filter"
        class="max-w-sm"
        placeholder="Search user records..."
      />
      <DialogNewUser />
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
              No results
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
