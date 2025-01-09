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

  useEffect(() => {
    if (location.pathname === "/") {
      redirectPipeline()
      return;
    } else if (location.pathname === "/login") {
      return;
    }
    loadUser();
  }, []);


  const loadUser = async () => {
    const res = await fetchRequest('/api/userinfo', {
      method: 'GET',
    });
    setUser(res);
    localStorage.setItem("userInfo", JSON.stringify(res));
  }

  const redirectPipeline = () => {
    navigate("/pipeline");
  };

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("userInfo");
    navigate('/login');
  }

  return (
    <Layout>
      <Header
        style={{
          position: "sticky",
          top: 0,
          zIndex: 1,
          width: "100%",
          display: "flex",
          alignItems: "center",
        }}
      >
        <Menu
          theme="dark"
          mode="horizontal"
          defaultSelectedKeys={[window.location.pathname.split('/')[1]]}
          onClick={handleMenuClick}
          items={[
            {
              key: "pipeline",
              label: "pipeline",
            },
            {
              key: "runner",
              label: "执行机器",
            },
          ]}
          style={{ flex: 1, minWidth: 0 }}
        />
        <Space>
            <Dropdown menu={{ items: [
              {
                key: '1',
                label: (
                  <a onClick={handleLogout}>
                    退出登录
                  </a>
                ),
              },
            ] }}>
              <Avatar src={user?.avatar} style={{ backgroundColor: "#743aed", position: "relative"}}>
                {user?.username.slice(0, 1)}
              </Avatar>
            </Dropdown>
          </Space>
      </Header>
      <Content style={{ padding: "20px 20px", height: "calc(100vh - 64px)" }}>
        <Outlet />
      </Content>
    </Layout>
  );
};

export default App;
