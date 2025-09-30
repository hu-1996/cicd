import React, { useState, useEffect } from "react";
import {
  Table,
  Tag,
  Button,
  message,
  Popconfirm,
  Tabs,
  Form,
  Modal,
  Input,
  Select,
} from "antd";
import type { TableProps, TabsProps } from "antd";
import { fetchRequest } from "../../utils/fetch";

interface User {
  id: number;
  username: string;
  nickname: string;
  roles: number[];
}

interface Role {
  id: number;
  name: string;
}

interface LoadDataParams {
  name?: string;
  page: number;
  page_size: number;
}

const IndexUser: React.FC = () => {
  const [data, setData] = useState<User[]>([]);
  const [roleData, setRoleData] = useState<Role[]>([]);
  const [roleMap, setRoleMap] = useState<Record<number, string>>({});

  //   const [form] = Form.useForm();
  const [queryParams, setQueryParams] = useState<LoadDataParams>({
    page: 1,
    page_size: 10,
  });
  const [total, setTotal] = useState(0);

  useEffect(() => {
    loadData();
    loadRoleData();
  }, [queryParams]);

  const deleteUser = async (userId: number) => {
    await fetchRequest("/api/delete_user/" + userId, {
      method: "DELETE",
    });
    loadData();
    message.success("已删除");
  };

  const deleteRole = async (roleId: number) => {
    await fetchRequest("/api/delete_role/" + roleId, {
      method: "DELETE",
    });
    message.success("已删除");
  };

  const columns: TableProps<User>["columns"] = [
    {
      title: "账号",
      dataIndex: "username",
      key: "username",
      render: (text) => <span>{text}</span>,
    },
    {
      title: "昵称",
      dataIndex: "nickname",
      key: "nickname",
      render: (text) => <span>{text || "-"}</span>,
    },
    {
      title: "角色",
      dataIndex: "roles",
      key: "roles",
      render: (roles) =>
        roles.map((roleId: number) => {
          return (
            <Tag key={roleId} bordered={false} color="processing">
              {roleMap[roleId]}
            </Tag>
          );
        }),
    },
    {
      title: "创建时间",
      dataIndex: "created_at",
      key: "created_at",
      width: 200,
    },
    {
      title: "更新时间",
      dataIndex: "updated_at",
      key: "updated_at",
      width: 200,
    },
    {
      title: "操作",
      key: "action",
      render: (_, record) => (
        <>
          <Button
            type="link"
            onClick={() => {
              setFormUser(record);
              form.setFieldsValue(record);
              setOpen(true);
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            onClick={() => {
              setFormUser(record);
              formChangePassword.resetFields();
              setOpenChangePassword(true);
            }}
          >
            更改密码
          </Button>
          <Popconfirm
            title="提示"
            description={`是否删除${record.nickname}?`}
            onConfirm={() => deleteUser(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger>
              删除
            </Button>
          </Popconfirm>
        </>
      ),
    },
  ];

  const roleColumns: TableProps<Role>["columns"] = [
    {
      title: "角色",
      dataIndex: "name",
      key: "name",
      render: (text) => <span>{text}</span>,
    },
    {
      title: "创建时间",
      dataIndex: "created_at",
      key: "created_at",
      width: 200,
    },
    {
      title: "操作",
      key: "action",
      render: (_, record) => (
        <>
          <Popconfirm
            title="提示"
            description={`是否删除${record.name}?`}
            onConfirm={() => deleteRole(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger>
              删除
            </Button>
          </Popconfirm>
        </>
      ),
    },
  ];

  const loadData = async () => {
    const params = new URLSearchParams(
      Object.entries(queryParams).reduce((acc, [key, value]) => {
        acc[key] = String(value);
        return acc;
      }, {} as Record<string, string>)
    ).toString();
    const data = await fetchRequest(`/api/list_user?` + params);
    setData(data.list);
    setTotal(data.total);
  };

  const loadRoleData = async () => {
    const data = await fetchRequest(`/api/list_role`);
    setRoleData(data.list);
    let rm: Record<number, string> = {};
    data.list.forEach((role: Role) => {
      rm[role.id] = role.name;
    });
    setRoleMap(rm);
  };

  const [form] = Form.useForm();
  const [formUser, setFormUser] = useState<User>();
  const [open, setOpen] = useState(false);
  const [openChangePassword, setOpenChangePassword] = useState(false);
  const [formChangePassword] = Form.useForm();
  const [openNewRole, setOpenNewRole] = useState(false);
  const [formNewRole] = Form.useForm();

  const onCreate = async (values: User) => {
    console.log("Received values of form: ", values);
    const url = formUser
      ? `/api/update_user/${formUser?.id}`
      : `/api/create_user`;
    await fetchRequest(url, {
      method: formUser ? "PUT" : "POST",
      body: JSON.stringify(values),
    });
    setOpen(false);
    loadData();
  };

  const onChangePassword = async (values: User) => {
    await fetchRequest(`/api/reset_password/${formUser?.id}`, {
      method: "PUT",
      body: JSON.stringify(values),
    });
    setOpenChangePassword(false);
  };

  const onCreateRole = async (values: Role) => {
    await fetchRequest('/api/create_role', {
      method: "POST",
      body: JSON.stringify(values),
    });
    setOpenNewRole(false);
    loadRoleData();
  };

  const items: TabsProps["items"] = [
    {
      key: "1",
      label: "用户列表",
      children: (
        <>
          <Button
            type="primary"
            onClick={() => {
              form.resetFields();
              setFormUser(undefined);
              setOpen(true);
            }}
          >
            添加用户
          </Button>
          <Table<User>
            columns={columns}
            dataSource={data}
            style={{ marginTop: 16 }}
            rowKey="id"
            pagination={{
              pageSize: queryParams.page_size,
              current: queryParams.page,
              total: total,
              onChange: (page, pageSize) => {
                setQueryParams({
                  ...queryParams,
                  page,
                  page_size: pageSize,
                });
              },
            }}
          />
          <Modal
            open={open}
            title={formUser ? "编辑用户" : "添加用户"}
            okText="保存"
            cancelText="取消"
            okButtonProps={{ autoFocus: true, htmlType: "submit" }}
            onCancel={() => setOpen(false)}
            destroyOnHidden
            modalRender={(dom) => (
              <Form
                layout="vertical"
                form={form}
                name="create_or_update_user"
                initialValues={{ modifier: "public" }}
                clearOnDestroy
                onFinish={(values) => onCreate(values)}
              >
                {dom}
              </Form>
            )}
          >
            <Form.Item name="username" label="账号" rules={[{ required: true, message: "请输入账号" }]}>
              <Input readOnly={!!formUser} placeholder="请输入账号" />
            </Form.Item>
            <Form.Item name="nickname" label="昵称" rules={[{ required: true, message: "请输入昵称" }]}>
              <Input placeholder="请输入昵称" />
            </Form.Item>
            {!formUser && (
              <Form.Item name="password" label="密码" rules={[{ required: true, message: "请输入密码" }]}>
                <Input.Password placeholder="请输入密码" />
              </Form.Item>
            )}
            <Form.Item name="roles" label="角色" rules={[{ required: true, message: "请选择角色" }]}>
              <Select
                mode="multiple"
                style={{ width: "100%" }}
                placeholder="请选择角色"
                options={roleData.map((role) => ({
                  key: role.id,
                  value: role.id,
                  label: role.name,
                }))}
              />
            </Form.Item>
          </Modal>
          <Modal
            open={openChangePassword}
            title={"更改" + formUser?.nickname + "的密码"}
            okText="保存"
            cancelText="取消"
            okButtonProps={{ autoFocus: true, htmlType: "submit" }}
            onCancel={() => setOpenChangePassword(false)}
            destroyOnHidden
            modalRender={(dom) => (
              <Form
                layout="vertical"
                form={formChangePassword}
                name="change_password"
                initialValues={{ modifier: "public" }}
                clearOnDestroy
                onFinish={(values) => onChangePassword(values)}
              >
                {dom}
              </Form>
            )}
          >
            <Form.Item name="password" label="密码" rules={[{ required: true, message: "请输入密码" }]}>
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
            <Form.Item name="password2" label="确认密码" rules={[{ required: true, message: "请输入确认密码" }]}>
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
          </Modal>
        </>
      ),
    },
    {
      key: "2",
      label: "角色列表",
      children: (
        <>
          <Button
            type="primary"
            onClick={() => {
              formNewRole.resetFields();
              setOpenNewRole(true);
            }}
          >
            添加角色
          </Button>
          <Table<Role>
            columns={roleColumns}
            dataSource={roleData}
            style={{ marginTop: 16 }}
            rowKey="id"
          />
          <Modal
            open={openNewRole}
            title={"添加角色"}
            okText="保存"
            cancelText="取消"
            okButtonProps={{ autoFocus: true, htmlType: "submit" }}
            onCancel={() => setOpenNewRole(false)}
            destroyOnHidden
            modalRender={(dom) => (
              <Form
                layout="vertical"
                form={formNewRole}
                name="new_role"
                initialValues={{ modifier: "public" }}
                clearOnDestroy
                onFinish={(values) => onCreateRole(values)}
              >
                {dom}
              </Form>
            )}
          >
            <Form.Item name="name" label="角色" rules={[{ required: true, message: "请输入角色" }]}>
              <Input placeholder="请输入角色" />
            </Form.Item>
          </Modal>
        </>
      ),
    },
  ];

  return (
    <>
      <Tabs defaultActiveKey="1" items={items} />
    </>
  );
};
export default IndexUser;
