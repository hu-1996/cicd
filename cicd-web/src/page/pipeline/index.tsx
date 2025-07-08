import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, Button, message, List, Popconfirm, Space } from "antd";
import {
  SettingOutlined,
  CaretRightOutlined,
  FolderOpenOutlined,
  DeleteOutlined,
  CopyOutlined,
} from "@ant-design/icons";
import Status from "./component/status";

import { fetchRequest } from "../../utils/fetch";

export default function Pipeline() {
  const navigate = useNavigate();

  const [pipelines, setPipelines] = useState([]);
  useEffect(() => {
    loadPipeline();
    const timer = setInterval(() => {
      loadPipeline();
    }, 5000);

    return () => {
      clearInterval(timer); // 清理定时器避免内存泄漏
    };
  }, []);

  const loadPipeline = async () => {
    const res = await fetchRequest("/api/list_pipeline", {
      method: "GET",
    });
    setPipelines(res || []);
  };

  const startJob = async (id: number) => {
    await fetchRequest("/api/start_job/" + id, {
      method: "POST",
    });
    message.success("启动成功");
    loadPipeline();
  };

  const toCreate = () => {
    navigate("/new_pipeline/pipeline");
  };

  const toSetting = () => {
    navigate("/setting");
  };

  const confirm = async (id: number) => {
    await fetchRequest("/api/delete_pipeline/" + id, {
      method: "DELETE",
    });
    message.success("删除成功");
    loadPipeline();
  };

  const copyPipeline = async (id: number) => {
    await fetchRequest("/api/copy_pipeline/" + id, {
      method: "POST",
    });
    message.success("复制成功");
    loadPipeline();
  };

  return (
    <div className="p-2">
      <Space>
        <Button type="primary" onClick={toCreate}>
          新建Pipeline
        </Button>
        <Button onClick={toSetting} icon={<SettingOutlined />}>
          设置
        </Button>
      </Space>
      {pipelines.map((item: any) => (
        <div key={item.group_name} className="mb-4 mt-4">
          <div className="text-lg font-bold pt-2 pb-2 pl-4 pr-4 bg-[#E8EEF0] text-[#3F82C9] rounded-t-md">
            {item.group_name || "Default"}
          </div>
          <div className="p-4 bg-white rounded-b-md">
            <List
              grid={{
                gutter: 16,
              }}
              dataSource={item.pipelines}
              renderItem={(item: any) => (
                <List.Item>
                  <Card
                    key={item.id}
                    title={item.name}
                    actions={[
                      <CaretRightOutlined
                        key={"start"}
                        onClick={() => startJob(item.id)}
                      />,
                      <FolderOpenOutlined
                        key={"history"}
                        onClick={() => navigate("/history?id=" + item.id)}
                      />,
                      <SettingOutlined
                        key={"setting"}
                        onClick={() =>
                          navigate("/new_pipeline/pipeline?id=" + item.id)
                        }
                      />,
                      <CopyOutlined
                        key={"copy"}
                        onClick={() => copyPipeline(item.id)}
                      />,
                      <Popconfirm
                        title="提示"
                        description={`是否删除${item.name}?`}
                        onConfirm={() => confirm(item.id)}
                        okText="确定"
                        cancelText="取消"
                      >
                        <DeleteOutlined key={"setting"} />
                      </Popconfirm>,
                    ]}
                    style={{ width: 300 }}
                  >
                    <div className="text-[12px]">Tag：{item.last_tag}</div>
                    <div className="text-[12px]">
                      最后更新时间：{item.last_update_at}
                    </div>
                    <div className="flex">
                      <Status steps={item.steps} />
                    </div>
                  </Card>
                </List.Item>
              )}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
