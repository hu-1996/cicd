import React, { useEffect, useState } from "react";
import { Layout, Menu, Space, Avatar, Dropdown } from "antd";
import { Outlet, useNavigate } from "react-router-dom";
import { fetchRequest } from "./utils/fetch";
const { Header, Content } = Layout;

const App: React.FC = () => {
  const navigate = useNavigate();
  const handleMenuClick = (e: any) => {
    navigate(e.key);
  };

  const [user, setUser] = useState<any>(null);
  const [items, setItems] = useState<any[]>([
    {
      key: "pipeline",
      label: "pipeline",
    },
    {
      key: "runner",
      label: "执行机器",
    },
  ]);

  useEffect(() => {
    if (location.pathname === "/") {
      redirectPipeline();
      return;
    } else if (location.pathname === "/login") {
      return;
    }
    loadUser();
  }, []);

  const loadUser = async () => {
    const res = await fetchRequest("/api/userinfo", {
      method: "GET",
    });
    setUser(res);
    localStorage.setItem("userinfo", JSON.stringify(res));
    if (res.is_admin) {
      setItems([...items, { key: "user", label: "用户" }]);
    }
  };

  const redirectPipeline = () => {
    navigate("/pipeline");
  };

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("userinfo");
    navigate("/login");
  };


  return (
    <Layout>
      <Header
        style={{
          display: "flex",
          alignItems: "center",
          
        }}
      >
        <Menu
          theme="dark"
          mode="horizontal"
          defaultSelectedKeys={[window.location.pathname.split("/")[1]]}
          onClick={handleMenuClick}
          items={items}
          style={{ flex: 1, minWidth: 0 }}
        />
        <Space>
          <Dropdown
            menu={{
              items: [
                {
                  key: "profile",
                  label: <a onClick={() => navigate("/profile")}>个人中心</a>,
                },
                {
                  type: 'divider',
                },
                {
                  key: "logout",
                  label: <a onClick={handleLogout}>退出登录</a>,
                },
              ],
            }}
          >
            <Avatar
              src={user?.avatar}
              style={{ backgroundColor: "#743aed", position: "relative" }}
            >
              {user?.nickname.slice(0, 1)}
            </Avatar>
          </Dropdown>
        </Space>
      </Header>
      <Content>
        <div
          style={{
            background: "#f8fbf8",
            padding: "10px 20px",
            borderRadius: 8,
            height: "calc(100vh - 64px)",
            overflow: "auto",
          }}
        >
          <Outlet />
        </div>
      </Content>
    </Layout>
  );
};

export default App;
