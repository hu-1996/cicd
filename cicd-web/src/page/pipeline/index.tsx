import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, Button, message, List, Popconfirm } from "antd";
import {
  SettingOutlined,
  CaretRightOutlined,
  FolderOpenOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import Status from "./component/status";

import { fetchRequest } from "../../utils/fetch";

export default function Pipeline() {
  const navigate = useNavigate();

  const [pipelines, setPipelines] = useState([]);
  useEffect(() => {
    loadPipeline();
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
  };

  const toCreate = () => {
    navigate("/new_pipeline/pipeline");
  };

  const confirm = async (id: number) => {
    await fetchRequest("/api/delete_pipeline/" + id, {
      method: "DELETE",
    });
    message.success("删除成功")
    loadPipeline();
  };

  return (
    <div className="p-2">
      <Button type="primary" className="mb-2" onClick={toCreate}>
        新建Pipeline
      </Button>
      <List
        grid={{
          gutter: 16,
        }}
        dataSource={pipelines}
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
                <Popconfirm
                  title="提示"
                  description={`是否删除${item.name}?`}
                  onConfirm={() => confirm(item.id)}
                  okText="确定"
                  cancelText="取消"
                >
                  <DeleteOutlined
                    key={"setting"}
                  />
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
  );
}