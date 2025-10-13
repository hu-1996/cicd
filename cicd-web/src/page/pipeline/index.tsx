import { useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Card, Button, message, List, Popconfirm, Space, Input } from "antd";
import {
  SettingOutlined,
  CaretRightOutlined,
  FolderOpenOutlined,
  DeleteOutlined,
  CopyOutlined,
} from "@ant-design/icons";
import Status from "./component/status";
import { colors, status, intro } from "./../../config/consts";

import { fetchRequest } from "../../utils/fetch";

export default function Pipeline() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const [originPipelines, setOriginPipelines] = useState<any[]>([]);
  const [pipelines, setPipelines] = useState<any[]>([]);
  const [search, setSearch] = useState<string>("");

  const q = searchParams.get("q");

  useMemo(() => {
    if (!search) {
      setPipelines(originPipelines);
      return;
    }
    let filtered: any[] = [];
    originPipelines.forEach((item: any) => {
      const pipelines = item.pipelines.filter((pipeline: any) => {
        return pipeline.name.includes(search);
      });
      if (pipelines.length > 0) {
        filtered.push({
          ...item,
          pipelines,
        });
      }
    });
    setPipelines(filtered);
  }, [search, originPipelines]);

  useEffect(() => {
    setSearch(q || "");
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
    setOriginPipelines(res || []);
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

  const onSearch = (e: any) => {
    setSearch(e.target.value);
    const url = new URL(window.location.href);
    if (e.target.value) {
      url.searchParams.set("q", e.target.value);
    } else {
      url.searchParams.delete("q");
    }

    // 使用history.pushState更新URL而不刷新页面
    history.pushState({}, "", url.toString());
  };

  return (
    <div className="p-2">
      <Space>
        <Input
          placeholder="搜索Pipeline"
          onChange={onSearch}
          value={search}
        ></Input>
        <Button type="primary" onClick={toCreate}>
          新建Pipeline
        </Button>
        <Button onClick={toSetting} icon={<SettingOutlined />}>
          设置
        </Button>
      </Space>
      <Space className="float-right">
        {Object.entries(status).map((item: any) => (
          <Space key={item[0]}>
            <div
              className="h-[16px] w-[40px] bg-[#afafb0] text-center leading-[16px] cursor-pointer"
              style={{
                backgroundColor: colors[item[0]],
              }}
            >
              {item[1]}
            </div>
            {intro[item[0]]}
          </Space>
        ))}
      </Space>
      {pipelines?.map((item: any) => (
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
                  <div style={{ minWidth: 300 }}>
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
                            navigate(
                              "/new_pipeline/pipeline?id=" +
                                item.id +
                                "&q=" +
                                search
                            )
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
                      style={{ width: "100%" }}
                    >
                      <div className="text-[12px]">Tag：{item.last_tag}</div>
                      <div className="text-[12px]">
                        最后更新时间：{item.last_update_at}
                      </div>
                      <div className="flex mt-1 min-h-[26px]">
                        <Status steps={item.steps} />
                      </div>
                    </Card>
                  </div>
                </List.Item>
              )}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
