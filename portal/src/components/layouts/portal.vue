<template>
    <el-container style="height: 100vh; border: 1px solid #eee">
        <el-header class="header">

        </el-header>

        <el-container>
            <el-aside width="200px" style="background-color: rgb(238, 241, 246)">
                <el-menu :router="true" class="totalUl">
                    <el-menu-item index="/portal/buckets"><i class="el-icon-message"></i>Bucket</el-menu-item>
                    <el-menu-item index="1"><i class="el-icon-message"></i>Other</el-menu-item>
                </el-menu>
            </el-aside>
            <el-main>
                <section class="content">
                    <transition
                            name="page"
                            mode="out-in">
                        <router-view></router-view>
                    </transition>
                </section>
            </el-main>
        </el-container>
    </el-container>
</template>

<script>
    export default {
        name: "portal"
    }
</script>

<style scoped>
    .header {
        color: rgba(255, 255, 255, 0.75);
        line-height: 60px;
        background-color: #24292e;
        text-align: center;
    }

    .header div {
        display: inline;
    }

    .totalUl {
        height: 100%;
    }
</style>