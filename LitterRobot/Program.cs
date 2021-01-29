using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using Microsoft.Extensions.Caching.Memory;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using LitterRobot.DataAccess;
using LitterRobot.Liasons;
using LitterRobot.Models.Shared;
using TwoMQTT;
using TwoMQTT.Extensions;
using TwoMQTT.Interfaces;
using TwoMQTT.Managers;
using System;

namespace LitterRobot
{
    class Program
    {
        static async Task Main(string[] args) => await ConsoleProgram<Resource, Command, SourceLiason, MQTTLiason>.
            ExecuteAsync(args,
                envs: new Dictionary<string, string>()
                {
                    {
                        $"{Models.Options.MQTTOpts.Section}:{nameof(Models.Options.MQTTOpts.TopicPrefix)}",
                        Models.Options.MQTTOpts.TopicPrefixDefault
                    },
                    {
                        $"{Models.Options.MQTTOpts.Section}:{nameof(Models.Options.MQTTOpts.DiscoveryName)}",
                        Models.Options.MQTTOpts.DiscoveryNameDefault
                    },
                },
                configureServices: (HostBuilderContext context, IServiceCollection services) =>
                {
                    services
                        .AddHttpClient()
                        .AddMemoryCache()
                        .AddOptions<Models.Options.SharedOpts>(Models.Options.SharedOpts.Section, context.Configuration)
                        .AddOptions<Models.Options.SourceOpts>(Models.Options.SourceOpts.Section, context.Configuration)
                        .AddOptions<TwoMQTT.Models.MQTTManagerOptions>(Models.Options.MQTTOpts.Section, context.Configuration)
                        .AddSingleton<IThrottleManager, ThrottleManager>(x =>
                        {
                            var opts = x.GetRequiredService<IOptions<Models.Options.SourceOpts>>();
                            return new ThrottleManager(opts.Value.PollingInterval);
                        })
                        .AddSingleton<ISourceDAO>(x =>
                        {
                            var logger = x.GetRequiredService<ILogger<SourceDAO>>();
                            var httpClientFactory = x.GetRequiredService<IHttpClientFactory>();
                            var cache = x.GetRequiredService<IMemoryCache>();
                            var opts = x.GetRequiredService<IOptions<Models.Options.SourceOpts>>();
                            return new SourceDAO(logger, httpClientFactory, cache, opts.Value.Login, opts.Value.Password);
                        });
                });
    }
}