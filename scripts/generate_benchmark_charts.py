from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np


ROOT = Path(__file__).resolve().parents[1]
OUTPUT_DIR = ROOT / "docs" / "images"

SCENARIOS = ["标量字段", "结构体值", "结构体指针", "错误字段", "动态参数"]
SERIES = {
    "zap": {
        "time": [295.2, 465.2, 502.3, 213.5, 484.7],
        "bytes": [192, 208, 208, 64, 417],
        "allocs": [1, 3, 3, 1, 3],
    },
    "slog": {
        "time": [600.3, 825.5, 860.9, 568.1, 728.6],
        "bytes": [0, 208, 208, 48, 32],
        "allocs": [0, 4, 4, 1, 2],
    },
    "log + JSON": {
        "time": [355.1, 387.7, 443.6, 304.0, 1088.0],
        "bytes": [272, 272, 280, 176, 866],
        "allocs": [4, 4, 5, 4, 16],
    },
}

COLORS = {
    "zap": "#4C78A8",
    "slog": "#F58518",
    "log + JSON": "#54A24B",
}


def configure_style() -> None:
    plt.rcParams.update(
        {
            "font.sans-serif": ["Microsoft YaHei", "SimHei", "DejaVu Sans"],
            "axes.unicode_minus": False,
            "figure.facecolor": "white",
            "axes.facecolor": "white",
            "axes.edgecolor": "#B8C0CC",
            "axes.labelcolor": "#273142",
            "xtick.color": "#273142",
            "ytick.color": "#273142",
            "text.color": "#1F2937",
            "axes.titleweight": "bold",
        }
    )


def grouped_bars(ax, metric: str, value_format: str) -> None:
    positions = np.arange(len(SCENARIOS))
    width = 0.24

    for index, (name, values) in enumerate(SERIES.items()):
        offset = (index - 1) * width
        bars = ax.bar(
            positions + offset,
            values[metric],
            width,
            label=name,
            color=COLORS[name],
        )
        labels = [value_format.format(value) for value in values[metric]]
        ax.bar_label(bars, labels=labels, padding=3, fontsize=9, color="#273142")

    ax.set_xticks(positions, SCENARIOS)
    ax.grid(axis="y", color="#E5E7EB", linewidth=0.8)
    ax.set_axisbelow(True)
    ax.spines[["top", "right"]].set_visible(False)


def generate_latency_chart() -> None:
    figure, axis = plt.subplots(figsize=(11, 6.2))
    grouped_bars(axis, "time", "{:.1f}")
    axis.set_title("五种日志场景的单次耗时中位数", fontsize=16, pad=16)
    axis.set_ylabel("耗时（ns/op）")
    axis.set_ylim(0, 1220)
    axis.legend(frameon=False, ncol=3, loc="upper left")
    figure.text(
        0.5,
        0.015,
        "数据来源：go test -run '^$' -bench . -benchmem -count=5；柱值为 5 次结果中位数",
        ha="center",
        fontsize=9,
        color="#5F6B7A",
    )
    figure.tight_layout(rect=(0, 0.05, 1, 1))
    figure.savefig(OUTPUT_DIR / "benchmark-latency.png", dpi=180, bbox_inches="tight")
    plt.close(figure)


def generate_memory_chart() -> None:
    figure, axes = plt.subplots(1, 2, figsize=(14, 6.2))

    grouped_bars(axes[0], "bytes", "{:.0f}")
    axes[0].set_title("单次堆内存分配量", fontsize=14, pad=14)
    axes[0].set_ylabel("分配量（B/op）")
    axes[0].set_ylim(0, 980)

    grouped_bars(axes[1], "allocs", "{:.0f}")
    axes[1].set_title("单次堆分配次数", fontsize=14, pad=14)
    axes[1].set_ylabel("分配次数（allocs/op）")
    axes[1].set_ylim(0, 18)

    for axis in axes:
        legend = axis.get_legend()
        if legend is not None:
            legend.remove()

    handles, labels = axes[0].get_legend_handles_labels()
    figure.legend(handles, labels, frameon=False, ncol=3, loc="upper center", bbox_to_anchor=(0.5, 0.98))
    figure.suptitle("五种日志场景的堆分配对比", fontsize=16, fontweight="bold", y=1.04)
    figure.text(
        0.5,
        0.015,
        "B/op 表示每次调用分配的堆字节数；allocs/op 表示每次调用发生的堆分配次数",
        ha="center",
        fontsize=9,
        color="#5F6B7A",
    )
    figure.tight_layout(rect=(0, 0.05, 1, 0.93))
    figure.savefig(OUTPUT_DIR / "benchmark-memory.png", dpi=180, bbox_inches="tight")
    plt.close(figure)


def main() -> None:
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    configure_style()
    generate_latency_chart()
    generate_memory_chart()


if __name__ == "__main__":
    main()
