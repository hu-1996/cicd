import React, { useEffect, useState } from "react";
import { List, Typography, Button, Space, Modal, Form, Input } from "antd";
import { fetchRequest } from "../../utils/fetch";
import { useNavigate } from "react-router-dom";

const { Title } = Typography;

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

const App: React.FC = () => {
  const navigate = useNavigate();
  const [user, setUser] = useState<any>(null);
  const [data, setData] = useState<any[]>([]);
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();
  const [openChangePassword, setOpenChangePassword] = useState(false);
  const [formChangePassword] = Form.useForm();

  useEffect(() => {
    loadUser();
  }, []);

  const loadUser = async () => {
    const res = await fetchRequest("/api/userinfo", {
      method: "GET",
    });
    setUser(res);
    localStorage.setItem("userinfo", JSON.stringify(res));

    const roleData = await fetchRequest(`/api/list_role`);
    const roleMap: Record<number, string> = {};
    roleData.list.forEach((role: Role) => {
      roleMap[role.id] = role.name;
    });
    setData([
      {
        title: "账号",
        description: res.username,
        readonly: true,
      },
      {
        title: "昵称",
        description: res.nickname,
      },
      {
        title: "密码",
        description: "********",
      },
      {
        title: "角色",
        description: res.roles
          .map((roleId: number) => roleMap[roleId])
          .join(", "),
        readonly: true,
      },
      {
        title: "注册时间",
        description: res.created_at,
        readonly: true,
      },
    ]);
  };

  const onUpdate = async (values: User) => {
    user.nickname = values.nickname;
    await fetchRequest(`/api/update_user/${user?.id}`, {
      method: "PUT",
      body: JSON.stringify(user),
    });
    setOpen(false);
    loadUser();
  };

  const onChangePassword = async (values: User) => {
    await fetchRequest(`/api/reset_password/${user?.id}`, {
      method: "PUT",
      body: JSON.stringify(values),
    });
    setOpenChangePassword(false);
    localStorage.removeItem("token");
    localStorage.removeItem("userinfo");
    navigate("/login");
  };

  return (
    <div className="bg-white p-4 rounded">
      <Title level={5}>个人信息</Title>
      <List
        itemLayout="horizontal"
        dataSource={data}
        split={false}
        renderItem={(item) => (
          <List.Item>
            <List.Item.Meta title={item.title} description={item.description} />
          </List.Item>
        )}
      />
      <Space>
        <Button type="primary" onClick={() => setOpen(true)}>
          修改个人信息
        </Button>
        <Button type="primary" onClick={() => setOpenChangePassword(true)}>
          更改密码
        </Button>
      </Space>
      <Modal
        open={open}
        title="修改个人信息"
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
            onFinish={(values) => onUpdate(values)}
          >
            {dom}
          </Form>
        )}
      >
        <Form.Item
          name="nickname"
          label="昵称"
          rules={[{ required: true, message: "请输入昵称" }]}
        >
          <Input placeholder="请输入昵称" />
        </Form.Item>
      </Modal>
      <Modal
        open={openChangePassword}
        title={"更改密码"}
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
        <Form.Item
          name="password"
          label="密码"
          rules={[{ required: true, message: "请输入密码" }]}
        >
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
        <Form.Item
          name="password2"
          label="确认密码"
          rules={[{ required: true, message: "请输入确认密码" }]}
        >
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
      </Modal>
    </div>
  );
};

export default App;
