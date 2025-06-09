import React, { useEffect, useState, useContext, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Table, Button, message, Space, Popconfirm } from "antd";
import type { DragEndEvent } from "@dnd-kit/core";
import { DndContext } from "@dnd-kit/core";
import type { SyntheticListenerMap } from "@dnd-kit/core/dist/hooks/utilities";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";
import {
  arrayMove,
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  SettingOutlined,
  HolderOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import type { TableColumnsType } from "antd";

import { fetchRequest } from "../../utils/fetch";

interface DataType {
  id: number;
  name: string;
  repository: string;
  branch: string;
  tag_template: string;
  last_tag: string;
}

interface RowContextProps {
  setActivatorNodeRef?: (element: HTMLElement | null) => void;
  listeners?: SyntheticListenerMap;
}

const RowContext = React.createContext<RowContextProps>({});

const DragHandle: React.FC = () => {
  const { setActivatorNodeRef, listeners } = useContext(RowContext);
  return (
    <Button
      type="text"
      size="small"
      icon={<HolderOutlined />}
      style={{ cursor: "move" }}
      ref={setActivatorNodeRef}
      {...listeners}
    />
  );
};

interface RowProps extends React.HTMLAttributes<HTMLTableRowElement> {
  "data-row-key": string;
}

const Row: React.FC<RowProps> = (props) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: props["data-row-key"] });

  const style: React.CSSProperties = {
    ...props.style,
    transform: CSS.Translate.toString(transform),
    transition,
    ...(isDragging ? { position: "relative", zIndex: 9999 } : {}),
  };

  const contextValue = useMemo<RowContextProps>(
    () => ({ setActivatorNodeRef, listeners }),
    [setActivatorNodeRef, listeners]
  );

  return (
    <RowContext.Provider value={contextValue}>
      <tr {...props} ref={setNodeRef} style={style} {...attributes} />
    </RowContext.Provider>
  );
};

export default function Setting() {
  const navigate = useNavigate();

  const columns: TableColumnsType<DataType> = [
    { key: "sort", align: "center", width: 80, render: () => <DragHandle /> },
    { title: "名称", dataIndex: "name" },
    { title: "仓库", dataIndex: "repository" },
    { title: "分支", dataIndex: "branch" },
    { title: "tag模板", dataIndex: "tag_template" },
    { title: "最新tag", dataIndex: "last_tag" },
    {
      title: "操作",
      dataIndex: "action",
      align: "center",
      width: 100,
      render: (_, record: DataType) => (
        <Space>
          <Popconfirm
            title="提示"
            description={`是否删除${record.name}?`}
            onConfirm={() => confirm(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <DeleteOutlined key={"setting"} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  interface PipelineGroup {
    group_name: string;
    pipelines: DataType[];
  }

  const [pipelines, setPipelines] = useState<PipelineGroup[]>([]);
  useEffect(() => {
    loadPipeline();
  }, []);

  const loadPipeline = async () => {
    const res = await fetchRequest("/api/list_pipeline", {
      method: "GET",
    });
    setPipelines(res || []);
  };

  const toCreate = () => {
    navigate("/new_pipeline/pipeline");
  };

  const confirm = async (id: number) => {
    await fetchRequest("/api/delete_pipeline/" + id, {
      method: "DELETE",
    });
    message.success("删除成功");
    loadPipeline();
  };

  const onDragEnd = async (groupName: string, { active, over }: DragEndEvent) => {
    if (active.id !== over?.id) {
      const oldPipelines = [...pipelines];
      const group = oldPipelines.find(
        (item: PipelineGroup) => item.group_name === groupName
      );
      if (group) {
        const activeIndex = group.pipelines?.findIndex(
          (record: DataType) => record.id === active?.id
        );
        const overIndex = group.pipelines?.findIndex(
          (record: DataType) => record.id === over?.id
        );
        if (activeIndex !== undefined && overIndex !== undefined) {
          const newPipelines = arrayMove(
            group.pipelines,
            activeIndex,
            overIndex
          );
          group.pipelines = newPipelines;

          await fetchRequest("/api/sort_pipeline", {
            method: "POST",
            body: JSON.stringify({ pipeline_ids: newPipelines.map((item: DataType) => item.id) }),
          });
        }
      }
      setPipelines(oldPipelines);
    }
  };

  return (
    <div className="p-2">
      <Space>
        <Button type="primary" onClick={toCreate}>
          新建Pipeline
        </Button>
        <Button onClick={toCreate} icon={<SettingOutlined />}>
          设置
        </Button>
      </Space>
      {pipelines.map((item: any) => (
        <div key={item.group_name} className="mb-4 mt-4">
          <div className="text-lg font-bold pt-2 pb-2 pl-4 pr-4 bg-[#E8EEF0] text-[#3F82C9] rounded-t-md">
            {item.group_name || "Default"}
          </div>
          <div className="p-4 bg-white rounded-b-md">
            <DndContext
              modifiers={[restrictToVerticalAxis]}
              onDragEnd={(event: DragEndEvent) =>
                onDragEnd(item.group_name, event)
              }
            >
              <SortableContext
                items={item.pipelines.map((i: any) => i.id)}
                strategy={verticalListSortingStrategy}
              >
                <Table<DataType>
                  rowKey="id"
                  components={{ body: { row: Row } }}
                  columns={columns}
                  dataSource={item.pipelines}
                  pagination={false}
                />
              </SortableContext>
            </DndContext>
          </div>
        </div>
      ))}
    </div>
  );
}
