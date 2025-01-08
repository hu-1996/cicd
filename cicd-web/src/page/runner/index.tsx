import React, { useState, useEffect } from "react";
import { Table, Switch, Tag, Button, message } from "antd";
import type { TableProps } from "antd";
import { fetchRequest } from "../../utils/fetch";

interface DataType {
  id: string;
  name: string;
  intro: string;
  logo_url: string;
  category: string;
  enabled: boolean;
  version_num: number;
  created_at: string;
}

interface LoadDataParams {
  name?: string;
  page: number;
  page_size: number;
  category?: string;
}

const IndexApplication: React.FC = () => {
  const [data, setData] = useState<DataType[]>([]);

  //   const [form] = Form.useForm();
  const [queryParams, setQueryParams] = useState<LoadDataParams>({
    page: 1,
    page_size: 10,
  });
  const [total, setTotal] = useState(0);

  useEffect(() => {
    loadData();
  }, [queryParams]);

  const onChange = async (record: any) => {
    await fetchRequest("/api/enable_runner/" + record.id, {
      method: "PUT",
    });
    loadData();
  };

  const setRunnerBusy = async (runnerId: string) => {
    await fetchRequest("/api/set_runner_busy/" + runnerId, {
      method: "PUT",
    });
    loadData();
    message.success("已设置为空闲");
  };

  const columns: TableProps<DataType>["columns"] = [
    {
      title: "名称",
      dataIndex: "name",
      key: "name",
      render: (text) => <span>{text}</span>,
    },
    {
      title: "标签",
      dataIndex: "labels",
      key: "labels",
      render: (labels) =>
        labels.map((label: any) => {
          return (
            <Tag key={label} bordered={false} color="processing">
              {label}
            </Tag>
          );
        }),
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      render: (obj) =>
        obj === "online" ? (
          <Tag bordered={false} color="processing">
            在线
          </Tag>
        ) : (
          <Tag bordered={false} color="gold">
            下线
          </Tag>
        ),
    },
    {
      title: "启用",
      dataIndex: "enable",
      key: "enable",
      render: (obj, record) => (
        <Switch
          defaultChecked={obj}
          onChange={() => onChange(record)}
        />
      ),
    },
    {
      title: "是否空闲",
      dataIndex: "busy",
      key: "busy",
      render: (obj) => (
        <Tag bordered={false} color={obj ? "gold" : "processing"}>
          {obj ? "忙碌" : "空闲"}
        </Tag>
      ),
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
        <Button type="link" onClick={() => setRunnerBusy(record.id)}>
          设置为空闲
        </Button>
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
    const data = await fetchRequest(`/api/list_runner?` + params);
    setData(data.list);
    setTotal(data.total);
  };

  //   const onFinish = (values: LoadDataParams) => {
  //     setQueryParams({
  //       ...queryParams,
  //       ...values,
  //       page: 1
  //     })
  //   };

  return (
    <>
      {/* <Form
        layout="inline"
        form={form}
        onFinish={onFinish}
        className='mb-4'
      >
        <Form.Item label="名称" name="name">
          <Input placeholder="请输入名称" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit">查询</Button>
        </Form.Item>
      </Form> */}
      <Table<DataType>
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
    </>
  );
};
export default IndexApplication;
