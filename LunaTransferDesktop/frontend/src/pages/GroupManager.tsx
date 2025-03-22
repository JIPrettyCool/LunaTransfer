import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Paper, 
  Button, 
  TextInput, 
  Table, 
  ActionIcon, 
  Menu, 
  Group, 
  Text, 
  Chip, 
  Avatar, 
  Badge, 
  Modal, 
  ScrollArea,
  Drawer,
  Title,
  Select,
  MultiSelect,
  Divider,
  Switch,
  Tooltip,
  LoadingOverlay,
  Box,
  rem
} from '@mantine/core';
import { IconPlus, IconDotsVertical, IconEdit, IconTrash, IconSearch, IconFilter, IconRefresh, IconUsers, IconLock, IconShield, IconUserPlus, IconDownload, IconUpload } from '@tabler/icons-react';
import { useDisclosure } from '@mantine/hooks';
import { notifications } from '@mantine/notifications';

interface GroupMember {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'member' | 'viewer';
  avatar?: string; 
}

interface Group {
  id: string;
  name: string;
  description: string;
  permissions: string[];
  members: GroupMember[];
  createdAt: string;
  isActive: boolean;
}

const mockGroups: Group[] = [
  {
    id: '1',
    name: 'Engineering',
    description: 'Software engineering department',
    permissions: ['upload', 'download', 'share'],
    members: [
      { id: '101', username: 'bendover', email: 'ben@example.com', role: 'admin', avatar: undefined },
      { id: '102', username: 'gokalaf', email: 'goktug@example.com', role: 'member', avatar: undefined }
    ],
    createdAt: '2025-03-01T12:00:00Z',
    isActive: true
  },
  {
    id: '2',
    name: 'Marketing',
    description: 'Marketing and PR team',
    permissions: ['download', 'view'],
    members: [
      { id: '103', username: 'joerogan', email: 'joe@example.com', role: 'admin', avatar: undefined },
      { id: '104', username: 'joegrahambell', email: 'grahambell@example.com', role: 'viewer', avatar: undefined }
    ],
    createdAt: '2025-03-05T09:30:00Z',
    isActive: true
  },
  {
    id: '3',
    name: 'Management',
    description: 'Company management and executives',
    permissions: ['upload', 'download', 'share', 'delete'],
    members: [
      { id: '105', username: 'robertdenuro', email: 'wiggle@example.com', role: 'admin', avatar: undefined }
    ],
    createdAt: '2025-02-15T15:45:00Z',
    isActive: true
  }
];

const availablePermissions = [
  { value: 'upload', label: 'Upload Files' },
  { value: 'download', label: 'Download Files' },
  { value: 'share', label: 'Share Files' },
  { value: 'delete', label: 'Delete Files' },
  { value: 'view', label: 'View Files' }
];

const GroupManager: React.FC = () => {
  const [groups, setGroups] = useState<Group[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [opened, { open, close }] = useDisclosure(false);
  const [confirmDeleteOpened, setConfirmDeleteOpened] = useState(false);
  const [groupToDelete, setGroupToDelete] = useState<string | null>(null);

  const [formData, setFormData] = useState<Partial<Group>>({
    name: '',
    description: '',
    permissions: [],
    isActive: true,
    members: []
  });

  const [membersSearch, setMembersSearch] = useState('');
  const [allUsers] = useState<GroupMember[]>([
    { id: '101', username: 'bendover', email: 'ben@example.com', role: 'admin' },
    { id: '102', username: 'gokalaf', email: 'goktug@example.com', role: 'member' },
    { id: '103', username: 'joerogan', email: 'joe@example.com', role: 'admin' },
    { id: '104', username: 'joegrahambell', email: 'grahambell@example.com', role: 'viewer' },
    { id: '105', username: 'robertdenuro', email: 'wiggle@example.com', role: 'admin' },
    { id: '106', username: 'bishbash', email: 'bosh@example.com', role: 'member' }
  ]);

  useEffect(() => {
    const timer = setTimeout(() => {
      setGroups(mockGroups);
      setIsLoading(false);
    }, 800);
    return () => clearTimeout(timer);
  }, []);

  const filteredGroups = groups.filter(group => 
    group.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
    group.description.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleOpenModal = (group: Group | null = null) => {
    if (group) {
      setFormData({...group});
      setSelectedGroup(group);
    } else {
      setFormData({
        name: '',
        description: '',
        permissions: [],
        isActive: true,
        members: []
      });
      setSelectedGroup(null);
    }
    open();
  };

  const handleSubmit = () => {
    setIsLoading(true);
    
    setTimeout(() => {
      if (selectedGroup) {
        setGroups(groups.map(g => g.id === selectedGroup.id ? { ...g, ...formData, id: g.id } as Group : g));
        notifications.show({
          title: 'Group Updated',
          message: `The group "${formData.name}" has been updated successfully.`,
          color: 'green',
        });
      } else {
        const newGroup: Group = {
          id: Date.now().toString(),
          name: formData.name || 'Unnamed Group',
          description: formData.description || '',
          permissions: formData.permissions || [],
          members: formData.members || [],
          createdAt: new Date().toISOString(),
          isActive: formData.isActive !== undefined ? formData.isActive : true
        };
        setGroups([...groups, newGroup]);
        notifications.show({
          title: 'Group Created',
          message: `The group "${newGroup.name}" has been created successfully.`,
          color: 'green',
        });
      }
      
      setIsLoading(false);
      close();
    }, 600);
  };

  const handleDeleteGroup = (groupId: string) => {
    setGroupToDelete(groupId);
    setConfirmDeleteOpened(true);
  };

  const confirmDelete = () => {
    if (groupToDelete) {
      setIsLoading(true);
      
      setTimeout(() => {
        const deletedGroup = groups.find(g => g.id === groupToDelete);
        setGroups(groups.filter(g => g.id !== groupToDelete));
        setConfirmDeleteOpened(false);
        setGroupToDelete(null);
        setIsLoading(false);
        
        if (deletedGroup) {
          notifications.show({
            title: 'Group Deleted',
            message: `The group "${deletedGroup.name}" has been deleted.`,
            color: 'red',
          });
        }
      }, 600);
    }
  };

  const handleMemberSelection = (memberId: string, role: 'admin' | 'member' | 'viewer') => {
    const user = allUsers.find(u => u.id === memberId);
    if (!user) return;

    const existingMember = formData.members?.find(m => m.id === memberId);
    
    if (existingMember) {
      setFormData({
        ...formData,
        members: formData.members?.map(m => 
          m.id === memberId ? { ...m, role } : m
        )
      });
    } else {
      setFormData({
        ...formData,
        members: [...(formData.members || []), { ...user, role }]
      });
    }
  };

  const handleRemoveMember = (memberId: string) => {
    setFormData({
      ...formData,
      members: formData.members?.filter(m => m.id !== memberId)
    });
  };

  const memberRoleOptions = [
    { value: 'admin', label: 'Admin' },
    { value: 'member', label: 'Member' },
    { value: 'viewer', label: 'Viewer' }
  ];

  return (
    <div className="p-4 md:p-8">
      <LoadingOverlay visible={isLoading} />
      
      <motion.div 
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <Group justify="space-between" mb={30}>
          <div>
            <Title order={2} className="flex items-center gap-2">
              <IconUsers size={28} className="text-indigo-500" />
              Group Management
            </Title>
            <Text c="dimmed" size="sm">
              Create and manage access groups for your organization
            </Text>
          </div>
          <Group>
            <Button 
              leftSection={<IconPlus size={16} />} 
              onClick={() => handleOpenModal()}
              className="bg-indigo-600 hover:bg-indigo-700"
              variant="filled"
            >
              Create Group
            </Button>
          </Group>
        </Group>
      </motion.div>

      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.3, delay: 0.1 }}
      >
        <Group justify="space-between" mb={20}>
          <TextInput
            placeholder="Search groups..."
            leftSection={<IconSearch size={16} />}
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.currentTarget.value)}
            className="w-full md:w-80"
          />
          <Group gap={8}>
            <Tooltip label="Refresh">
              <ActionIcon 
                color="gray" 
                variant="subtle"
                onClick={() => {
                  setIsLoading(true);
                  setTimeout(() => setIsLoading(false), 600);
                }}
              >
                <IconRefresh size={18} />
              </ActionIcon>
            </Tooltip>
            <Menu position="bottom-end" shadow="md">
              <Menu.Target>
                <Tooltip label="Filter">
                  <ActionIcon color="gray" variant="subtle">
                    <IconFilter size={18} />
                  </ActionIcon>
                </Tooltip>
              </Menu.Target>
              <Menu.Dropdown>
                <Menu.Label>Filter Groups</Menu.Label>
                <Menu.Item>Show Active Only</Menu.Item>
                <Menu.Item>Sort by Name</Menu.Item>
                <Menu.Item>Sort by Created Date</Menu.Item>
                <Menu.Divider />
                <Menu.Item>Reset Filters</Menu.Item>
              </Menu.Dropdown>
            </Menu>
          </Group>
        </Group>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, delay: 0.2 }}
        className="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden"
      >
        <ScrollArea>
          <Table striped highlightOnHover className="min-w-full">
            <Table.Thead className="bg-gray-50 dark:bg-gray-700">
              <Table.Tr>
                <Table.Th>Name</Table.Th>
                <Table.Th>Description</Table.Th>
                <Table.Th>Permissions</Table.Th>
                <Table.Th>Members</Table.Th>
                <Table.Th>Status</Table.Th>
                <Table.Th>Actions</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              <AnimatePresence>
                {filteredGroups.length === 0 ? (
                  <Table.Tr>
                    <Table.Td colSpan={6}>
                      <div className="text-center py-6 text-gray-500">
                        {searchTerm ? 'No groups match your search' : 'No groups have been created yet'}
                      </div>
                    </Table.Td>
                  </Table.Tr>
                ) : (
                  filteredGroups.map((group) => (
                    <motion.tr
                      key={group.id}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="hover:bg-gray-50 dark:hover:bg-gray-700"
                    >
                      <Table.Td className="whitespace-nowrap">
                        <Group gap="sm">
                          <div className="w-8 h-8 rounded-full bg-indigo-100 dark:bg-indigo-900 flex items-center justify-center text-indigo-600 dark:text-indigo-300 font-medium">
                            {group.name.charAt(0)}
                          </div>
                          <div>
                            <Text fw={600}>{group.name}</Text>
                            <Text size="xs" c="dimmed">Created {new Date(group.createdAt).toLocaleDateString()}</Text>
                          </div>
                        </Group>
                      </Table.Td>
                      <Table.Td>{group.description}</Table.Td>
                      <Table.Td>
                        <Group gap={4}>
                          {group.permissions.slice(0, 2).map(permission => (
                            <Badge 
                              key={permission} 
                              size="sm"
                              color={
                                permission === 'delete' ? 'red' : 
                                permission === 'share' ? 'blue' : 
                                permission === 'upload' ? 'green' : 
                                'gray'
                              }
                            >
                              {permission}
                            </Badge>
                          ))}
                          {group.permissions.length > 2 && (
                            <Tooltip label={group.permissions.slice(2).join(', ')}>
                              <Badge size="sm" color="gray">+{group.permissions.length - 2}</Badge>
                            </Tooltip>
                          )}
                        </Group>
                      </Table.Td>
                      <Table.Td>
                        <Group gap={4}>
                          <Avatar.Group>
                            {group.members.slice(0, 3).map(member => (
                              <Tooltip key={member.id} label={member.username}>
                                <Avatar 
                                  size="sm" 
                                  radius="xl" 
                                  src={member.avatar}
                                  color={
                                    member.role === 'admin' ? 'indigo' : 
                                    member.role === 'member' ? 'blue' : 
                                    'gray'
                                  }
                                >
                                  {member.username.charAt(0).toUpperCase()}
                                </Avatar>
                              </Tooltip>
                            ))}
                            {group.members.length > 3 && (
                              <Avatar size="sm" radius="xl">
                                +{group.members.length - 3}
                              </Avatar>
                            )}
                          </Avatar.Group>
                        </Group>
                      </Table.Td>
                      <Table.Td>
                        <Badge 
                          color={group.isActive ? 'green' : 'gray'}
                          variant="light"
                        >
                          {group.isActive ? 'Active' : 'Inactive'}
                        </Badge>
                      </Table.Td>
                      <Table.Td>
                        <Group gap={4}>
                          <Tooltip label="Edit Group">
                            <ActionIcon 
                              color="blue" 
                              variant="subtle"
                              onClick={() => handleOpenModal(group)}
                            >
                              <IconEdit size={16} />
                            </ActionIcon>
                          </Tooltip>
                          <Tooltip label="Delete Group">
                            <ActionIcon 
                              color="red" 
                              variant="subtle"
                              onClick={() => handleDeleteGroup(group.id)}
                            >
                              <IconTrash size={16} />
                            </ActionIcon>
                          </Tooltip>
                        </Group>
                      </Table.Td>
                    </motion.tr>
                  ))
                )}
              </AnimatePresence>
            </Table.Tbody>
          </Table>
        </ScrollArea>
      </motion.div>

      <Drawer
        opened={opened}
        onClose={close}
        padding="xl"
        size="xl"
        position="right"
        title={
          <Title order={3}>
            {selectedGroup ? `Edit Group: ${selectedGroup.name}` : 'Create New Group'}
          </Title>
        }
      >
        <div className="space-y-6">
          <TextInput
            label="Group Name"
            placeholder="Enter group name"
            required
            value={formData.name || ''}
            onChange={(e) => setFormData({...formData, name: e.target.value})}
          />
          
          <TextInput
            label="Description"
            placeholder="Describe the purpose of this group"
            value={formData.description || ''}
            onChange={(e) => setFormData({...formData, description: e.target.value})}
          />
          
          <MultiSelect
            label="Permissions"
            placeholder="Select permissions"
            data={availablePermissions}
            value={formData.permissions || []}
            onChange={(values) => setFormData({...formData, permissions: values})}
          />

          <Switch
            label="Active"
            checked={formData.isActive}
            onChange={(e) => setFormData({...formData, isActive: e.currentTarget.checked})}
          />

          <Divider my="md" label="Members" labelPosition="center" />

          <Group justify="space-between" mb={10}>
            <TextInput
              placeholder="Search users..."
              value={membersSearch}
              onChange={(e) => setMembersSearch(e.currentTarget.value)}
              leftSection={<IconSearch size={16} />}
              className="flex-1"
            />
          </Group>

          <Text fw={600} className="mt-4 mb-2">Group Members</Text>
          <ScrollArea style={{ height: 200 }}>
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>User</Table.Th>
                  <Table.Th>Role</Table.Th>
                  <Table.Th>Actions</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {formData.members && formData.members.length > 0 ? formData.members.map(member => (
                  <Table.Tr key={member.id}>
                    <Table.Td>
                      <Group gap="sm">
                        <Avatar size="sm" radius="xl">{member.username.charAt(0).toUpperCase()}</Avatar>
                        <div>
                          <Text fw={500}>{member.username}</Text>
                          <Text size="xs" c="dimmed">{member.email}</Text>
                        </div>
                      </Group>
                    </Table.Td>
                    <Table.Td>
                      <Select
                        value={member.role}
                        onChange={(value) => handleMemberSelection(member.id, value as 'admin' | 'member' | 'viewer')}
                        data={memberRoleOptions}
                        size="xs"
                        className="w-24"
                      />
                    </Table.Td>
                    <Table.Td>
                      <ActionIcon color="red" variant="subtle" onClick={() => handleRemoveMember(member.id)}>
                        <IconTrash size={16} />
                      </ActionIcon>
                    </Table.Td>
                  </Table.Tr>
                )) : (
                  <Table.Tr>
                    <Table.Td colSpan={3} className="text-center py-4 text-gray-500">
                      No members added yet
                    </Table.Td>
                  </Table.Tr>
                )}
              </Table.Tbody>
            </Table>
          </ScrollArea>

          <Text fw={600} className="mt-6 mb-2">Available Users</Text>
          <ScrollArea style={{ height: 200 }}>
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>User</Table.Th>
                  <Table.Th>Add As</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {allUsers
                  .filter(user => !formData.members?.some(m => m.id === user.id))
                  .filter(user => 
                    user.username.toLowerCase().includes(membersSearch.toLowerCase()) ||
                    user.email.toLowerCase().includes(membersSearch.toLowerCase())
                  )
                  .map(user => (
                    <Table.Tr key={user.id}>
                      <Table.Td>
                        <Group gap="sm">
                          <Avatar size="sm" radius="xl">{user.username.charAt(0).toUpperCase()}</Avatar>
                          <div>
                            <Text fw={500}>{user.username}</Text>
                            <Text size="xs" c="dimmed">{user.email}</Text>
                          </div>
                        </Group>
                      </Table.Td>
                      <Table.Td>
                        <Group gap="xs">
                          <Tooltip label="Add as Admin">
                            <ActionIcon 
                              color="indigo" 
                              variant="light"
                              onClick={() => handleMemberSelection(user.id, 'admin')}
                            >
                              <IconShield size={16} />
                            </ActionIcon>
                          </Tooltip>
                          <Tooltip label="Add as Member">
                            <ActionIcon 
                              color="blue" 
                              variant="light"
                              onClick={() => handleMemberSelection(user.id, 'member')}
                            >
                              <IconUsers size={16} />
                            </ActionIcon>
                          </Tooltip>
                          <Tooltip label="Add as Viewer">
                            <ActionIcon 
                              color="gray" 
                              variant="light"
                              onClick={() => handleMemberSelection(user.id, 'viewer')}
                            >
                              <IconLock size={16} />
                            </ActionIcon>
                          </Tooltip>
                        </Group>
                      </Table.Td>
                    </Table.Tr>
                  ))}
              </Table.Tbody>
            </Table>
          </ScrollArea>

          <Group justify="space-between" mt="xl">
            <Button variant="outline" onClick={close}>Cancel</Button>
            <Button 
              onClick={handleSubmit}
              disabled={!formData.name}
              className="bg-indigo-600 hover:bg-indigo-700"
            >
              {selectedGroup ? 'Update Group' : 'Create Group'}
            </Button>
          </Group>
        </div>
      </Drawer>

      <Modal
        opened={confirmDeleteOpened}
        onClose={() => setConfirmDeleteOpened(false)}
        title="Confirm Deletion"
        centered
      >
        <Text size="sm">
          Are you sure you want to delete this group? This action cannot be undone.
        </Text>
        <Group justify="flex-end" mt="md">
          <Button variant="outline" onClick={() => setConfirmDeleteOpened(false)}>
            Cancel
          </Button>
          <Button color="red" onClick={confirmDelete}>
            Delete
          </Button>
        </Group>
      </Modal>
    </div>
  );
};

export default GroupManager;