# AWS2.1
AWS2.x is a plugin for AIM2.x which is a hot-plug interface ,use this framework can trade the information in low lantency with websocket

# AWS是什么
AWS2是一个转发器，全称AWebSocket，这是AIM1（TuuzIM v5）的前置程序，通过AWS2可以实现功能层热更新等功能

# 为什么立项
AWS1.0在使用中，已经解决了性能问题，在2.0的时候主要为了实现热插拔，基于HTTP的数据交互也可以实现后端热插拔

# 性能&稳定性
理论上来说PHP+Swoole的性能是比Go好的，但是实际使用的时候，在量大的时候，
PHPswoole程序过一段时间就会出现卡死的问题，Go程序没有类似问题，
目前稳定运行1个月了，没有出现卡死等问题

稳定性方面主要是热更新程序的时候，所有连入端都不会掉，无闪断，后端程序可以平顺升级


# AWS vs AIM 架构设计

这次的架构方案还是沿用MChat（PHP版）v2的架构设计，因为Go特性，所以在使用协程后，大群效能可以提高1200%左右

# AWS-TCP技术

本版不支持TCP技术，TCP热插拔技术将会使用GRPC方案并在AWS3上线，2.0为了保证平滑升级，目前维持使用websocket技术
