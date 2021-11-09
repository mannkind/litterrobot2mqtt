using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using LitterRobot.DataAccess;
using LitterRobot.Liasons;
using LitterRobot.Models.Options;
using LitterRobot.Models.Shared;
using Microsoft.Extensions.Caching.Memory;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using TwoMQTT;
using TwoMQTT.Extensions;
using TwoMQTT.Interfaces;
using TwoMQTT.Managers;

await ConsoleProgram<Resource, Command, SourceLiason, MQTTLiason>.
    ExecuteAsync(args,
        envs: new Dictionary<string, string>()
        {
            {
                $"{MQTTOpts.Section}:{nameof(MQTTOpts.TopicPrefix)}",
                MQTTOpts.TopicPrefixDefault
            },
            {
                $"{MQTTOpts.Section}:{nameof(MQTTOpts.DiscoveryName)}",
                MQTTOpts.DiscoveryNameDefault
            },
        },
        configureServices: (HostBuilderContext context, IServiceCollection services) =>
        {
            services
                .AddHttpClient()
                .AddMemoryCache()
                .AddOptions<SharedOpts>(SharedOpts.Section, context.Configuration)
                .AddOptions<SourceOpts>(SourceOpts.Section, context.Configuration)
                .AddOptions<TwoMQTT.Models.MQTTManagerOptions>(MQTTOpts.Section, context.Configuration)
                .AddSingleton<IThrottleManager, ThrottleManager>(x =>
                {
                    var opts = x.GetRequiredService<IOptions<SourceOpts>>();
                    return new ThrottleManager(opts.Value.PollingInterval);
                })
                .AddSingleton<ISourceDAO>(x =>
                {
                    var logger = x.GetRequiredService<ILogger<SourceDAO>>();
                    var httpClientFactory = x.GetRequiredService<IHttpClientFactory>();
                    var cache = x.GetRequiredService<IMemoryCache>();
                    var opts = x.GetRequiredService<IOptions<SourceOpts>>();
                    return new SourceDAO(logger, httpClientFactory, cache, opts.Value.Login, opts.Value.Password);
                });
        });